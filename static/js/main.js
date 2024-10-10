const ACTION_STATUS = 0;
const ACTION_RESULT = 1;
const STATUS_INIT = 1;
const STATUS_FETCHING = 2;
const STATUS_ANALYZING = 3;
const STATUS_DONE = 4;

/**
 * @param {number} ms
 * @returns {Promise<void>}
 */
const wait = (ms) => new Promise((resolve) => setTimeout(resolve, ms));

/**
 *  POST "/"
 *	@typedef {Object} AnalyzeResponse
 *	@property {string} id
 *	@property {boolean} error
 *	@property {string} error_message
 */

/**
 * GET "/task/:id/:action"
 * @typedef {Object} TaskStatus
 * @property {number} task_status
 * @property {boolean} task_done
 * @property {boolean} task_error
 * @property {string} task_error_message
 */

class Client {
  /**
   *  @param {FormData} formData
   *  @returns {Promise<AnalyzeResponse | string>}
   */
  createTask(formData) {
    return fetch("/api/task", {
      method: "POST",
      body: new URLSearchParams(formData),
      headers: {
        Accept: "application/json",
        "Content-Type": "application/x-www-form-urlencoded",
      },
    }).then((response) => {
      const contentType = response.headers.get("content-type");
      const xCache = response.headers.get("X-Cache");
      const cached = xCache && xCache == "HIT";

      if (contentType && contentType.includes("text/html") && cached) {
        return response.text();
      }

      return response.json();
    });
  }

  /**
   * @param {taskID} string
   * @returns {Promise<TaskStatus>}
   */
  getTaskStatus(taskID) {
    return fetch(`/api/task/${taskID}/${ACTION_STATUS}`, {
      method: "GET",
      Accept: "application/json",
    }).then((response) => response.json());
  }

  /**
   *	@param {taskID} string
   *	@returns {Promise<string>}
   */
  getTaskResult(taskID) {
    return fetch(`/api/task/${taskID}/${ACTION_RESULT}`, {
      method: "GET",
      Accept: "text/html",
    }).then((response) => response.text());
  }
}

const client = new Client();

class App {
  /** @type {ReturnType<typeof setInterval>} */
  #fetchStatusInterval;
  #taskID = "";
  /** @type {AnalyzeForm} */
  #form = null;
  /** @type {Content} */
  #content = null;
  #hasAccess = true;

  constructor() {
    this.#content = new Content();

    this.#form = new AnalyzeForm({
      onSubmit: (e) => this.createTask(e),
    });
  }

  /**
   * @param {string} id
   */
  set taskId(id) {
    this.#taskID = id;
  }

  /**
   * @param {SubmitEvent} event
   * @returns {void}
   */
  createTask(event) {
    event.preventDefault();
    this.#form.disable = true;

    if (!this.#hasAccess) {
      return Promise.resolve();
    }

    this.#hasAccess = false;

    /**
     * @param {AnalyzeResponse | string} data
     */
    const handle = (data) => {
      if (typeof data == "string") {
        this.#content.renderAnalysisResult(data).then(() => {
          this.#hasAccess = true;
          this.#form.disable = false;
        });
        return;
      }

      if (data.error) {
        this.#content.renderError(data.error_message);
        this.#form.disable = false;
        this.#hasAccess = true;
        return;
      }

      this.#hasAccess = false;
      this.taskId = data.id;
      this.getTaskStatus();
      this.#fetchStatusInterval = setInterval(() => this.getTaskStatus(), 300);
    };

    const formData = new FormData(this.#form.element[0]);
    this.#content.renderStatus("STATUS_INIT");

    client
      .createTask(formData)
      .then(handle)
      .catch(() => this.#content.renderError("INTERNAL_ERROR"));
  }

  getTaskStatus() {
    /**
     * @param {Partial<TaskStatus>} data
     * @returns {void}
     */
    const handleData = (data) => {
      this.#content.renderStatus(data.task_status);

      if (data.task_done) {
        clearInterval(this.#fetchStatusInterval);

        if (data.task_error) {
          this.#form.disable = false;
          this.#hasAccess = false;
          this.#content.renderError(data.task_error_message);
          return;
        }

        this.getTaskResult();
      }
    };

    const handleException = () => {
      this.#content.renderError("INTERNAL_ERROR");
      clearInterval(this.#fetchStatusInterval);
    };

    return client
      .getTaskStatus(this.#taskID)
      .then(handleData)
      .catch(handleException);
  }

  /**
   * @returns {Promise<void>}
   */
  getTaskResult() {
    const handle = (data) => {
      this.#content.renderAnalysisResult(data).then(() => {
        this.#hasAccess = true;
        this.#form.disable = false;
      });
    };

    return client
      .getTaskResult(this.#taskID)
      .then(handle)
      .catch(() => this.#content.renderError("INTERNAL_ERROR"));
  }
}

/**
 *	@typedef {Object} RepoFormProps
 *	@property {() => void} onSubmit
 */

class AnalyzeForm {
  /**
   *	@type {RepoFormProps}
   */
  props = {};
  /** @type {JQuery<HTMLFormElement>} */
  #elem = $("#form");

  get element() {
    return this.#elem;
  }

  constructor(props) {
    this.props = props;
    this.#onReady();
  }

  #onReady() {
    this.#elem.on("submit", (e) => this.props.onSubmit(e));
    $("#btn-option").on("click", () => this.toggleOptionGroup());

    $("#option-group-file")
      .find(".input-option")
      .attr("placeholder", this.randomPlaceholder("file"));

    $("#option-group-dir")
      .find(".input-option")
      .attr("placeholder", this.randomPlaceholder("dir"));

    $("#option-group-file")
      .find(".btn-remove")
      .on("click", () => {
        this.removeOption($("#option-group-file").find(".btn-remove"));
      });

    $("#option-group-dir")
      .find(".btn-remove")
      .on("click", () => {
        this.removeOption($("#option-group-dir").find(".btn-remove"));
      });

    $("#btn-add-file-pattern").on("click", () => this.addOption("file"));
    $("#btn-add-dir-pattern").on("click", () => this.addOption("dir"));
  }

  /**
   * @typedef {"dir" | "file"} PatternGroup
   */

  /**
   * @param {PatternGroup} patternGroup
   */
  addOption(patternGroup) {
    const groupId = `#option-group-${patternGroup}`;
    const $group = $(groupId);
    const $item = $group.children().first().clone();
    const $btnRemove = $item.find(".btn-remove");

    $item
      .find(".input-text")
      .attr("placeholder", this.randomPlaceholder(patternGroup))
      .val("");

    $btnRemove.off("click").on("click", (e) => {
      this.removeOption($(e.currentTarget));
    });

    $item.appendTo($group);
    // $group.append($item);
  }

  /**
   * @param {JQuery<HTMLButtonElement>} $btn
   */
  removeOption($btn) {
    const $item = $btn.closest(".option-item").first();
    const $group = $btn.closest(".option-group");
    const $items = $group.children();

    if ($items.length > 1) {
      $item.remove();
    } else {
      $item.find(".input-text").val("");
    }
  }

  /**
   * @param {PatternGroup} patterGroup
   * @returns {string}
   */
  randomPlaceholder(patterGroup) {
    const placeholders = {
      dir: ["node_modules", "dist", ".idea", "__pycache__"],
      file: ["package*.json", "*.log", "README.md", "*.py"],
    }[patterGroup];

    return placeholders[Math.floor(Math.random() * placeholders.length)] || "";
  }

  toggleOptionGroup() {
    $("#options").toggleClass("hidden");
  }

  /**
   * @param {boolean} value
   */
  set disable(value) {
    $("button[type=submit]").prop("disabled", value);
  }
}

class Content {
  #elem = $("#content");
  #error = $("#error");

  constructor() {}

  get element() {
    return this.#elem;
  }

  /**
   *	@param {string} error
   */
  renderError(error) {
    this.#elem.empty();
    $("#form").removeClass("hidden");
    this.#error.removeClass("hidden").html(`
   <div class="error">
      <div class="error-title">
        <svg
          xmlns="http://www.w3.org/2000/svg"
          height="24px"
				  width="24px"
          viewBox="0 0 24 24"
        >
          <path d="M0 0h24v24H0V0z" fill="none" />
          <path d="M12 5.99L19.53 19H4.47L12 5.99M12 2L1 21h22L12 2zm1 14h-2v2h2v-2zm0-6h-2v4h2v-4z" />
        </svg>
        <h3>ERROR</h3>
      </div>
      <p>${error}</p>
    </div>
`);
  }

  /**
   * @param {number | undefined} status
   * @returns {void}
   */
  renderStatus(status) {
    this.#error.addClass("hidden");
    $("#form").addClass("hidden");

    const message =
      {
        1: "Sending...",
        2: "Fetching Repository...",
        3: "Analyzing Repository...",
        4: "Done.",
      }[status] || "Loading...";

    if (!this.#elem.find("#status-bar").length) {
      this.#elem.html(`
			<div id="status-bar" class="status">
				<h3 class="status-title">${message}</h3>
        <progress max="4" value=${status}></progress>
      </div>`);
      return;
    }

    this.#elem.find(".status-title").text(message);
    this.#elem.find("progress").val(status);
  }

  /**
   * @param {string} result
   * @returns {Promise<void>}
   */
  renderAnalysisResult(result) {
    this.#error.addClass("hidden");

    $("head").append(
      $("<link>")
        .attr("href", "css/table.css?v=" + new Date().getTime())
        .attr("rel", "stylesheet"),
    );

    return wait(1000)
      .then(() => {
        this.#elem.html(result);
        $("#form").removeClass("hidden");
      })
      .then(() => {
        $("html").animate({ scrollTop: $("#repo-table").offset().top }, 350);
      });
  }
}

$(() => {
  new App();
});
