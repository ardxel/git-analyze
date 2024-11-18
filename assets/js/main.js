import { client } from "./client.js";

const STATUS_INIT = 1;
const STATUS_FETCHING = 2;
const STATUS_ANALYZING = 3;
const STATUS_DONE = 4;
/**

 * @param {number} ms
 * @returns {Promise<void>}
 */
const wait = (ms) => new Promise((resolve) => setTimeout(resolve, ms));

const svgIcons = {
  error: `<svg
          xmlns="http://www.w3.org/2000/svg"
          height="24px"
				  width="24px"
          viewBox="0 0 24 24">
          <path d="M0 0h24v24H0V0z" fill="none" />
          <path d="M12 5.99L19.53 19H4.47L12 5.99M12 2L1 21h22L12 2zm1 14h-2v2h2v-2zm0-6h-2v4h2v-4z" />
        </svg>`,
  arrowUp: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 384 512"><path d="M214.6 41.4c-12.5-12.5-32.8-12.5-45.3 0l-160 160c-12.5 12.5-12.5 32.8 0 45.3s32.8 12.5 45.3 0L160 141.2 160 448c0 17.7 14.3 32 32 32s32-14.3 32-32l0-306.7L329.4 246.6c12.5 12.5 32.8 12.5 45.3 0s12.5-32.8 0-45.3l-160-160z"/></svg>`,
  arrowDown: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 384 512"><path d="M169.4 470.6c12.5 12.5 32.8 12.5 45.3 0l160-160c12.5-12.5 12.5-32.8 0-45.3s-32.8-12.5-45.3 0L224 370.8 224 64c0-17.7-14.3-32-32-32s-32 14.3-32 32l0 306.7L54.6 265.4c-12.5-12.5-32.8-12.5-45.3 0s-12.5 32.8 0 45.3l160 160z"/></svg>`,
};

class App {
  /** @type {ReturnType<typeof setInterval>} */
  #fetchStatusInterval;

  #taskID = "";

  /** @type {AnalyzeForm} */
  #form = null;

  /** @type {ResponseLayout} */
  #resLayout = null;

  #hasAccess = true;

  /** @type {import("./client.js").TaskResult}  */
  #resData = {};

  constructor() {
    this.#resLayout = new ResponseLayout();

    this.#form = new AnalyzeForm({
      onSubmit: (e) => this.createTask(e),
    });
  }

  get resLayout() {
    return this.#resLayout;
  }

  getCSS() {
    $("head").append($("<link>").attr("href", "css/table.css").attr("rel", "stylesheet"));
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
     * @param {import("./client.js").TaskInitResponse | string} data
     */
    const handle = (data) => {
      if (typeof data == "string") {
        this.#resLayout.renderResutlHTML(data).then(() => {
          this.#hasAccess = true;
          this.#form.disable = false;
        });
        return;
      }

      if (typeof data === "object" && data.error) {
        this.#resLayout.renderError(data.error_message);
        this.#form.disable = false;
        this.#hasAccess = true;
        return;
      }

      this.#hasAccess = false;
      this.taskId = data.id;
      this.getTaskStatus();
      this.#fetchStatusInterval = setInterval(() => this.getTaskStatus(), 300);
    };

    /**
     *	@callback TypeFnGet
     *	@param {keyof FormDataTsk} name
     *	@returns {FormDataTsk[name]}
     */

    /**
     *	@typedef {Object} FormDataTsk
     *	@property {string} repo_owner
     *	@property {string} repo_name
     *	@property {string} repo_url
     *	@property {Array<string>} exclude_file_patterns
     *	@property {Array<string>} exclude_dir_patterns
     *	@property {TypeFnGet} get
     */

    /** @type {FormDataTsk} */
    const formData = new FormData(this.#form.element[0]);

    if (window.gtag) {
      let repoUrl = formData.get("repo_url");
      let repoOwner = formData.get("repo_owner");
      let repoName = formData.get("repo_name");

      if ((!repoOwner || !repoName) && repoUrl) {
        try {
          const url = new URL(repoUrl);
          const parts = url.pathname.split("/");
          repoOwner = parts[1] || "invalid";
          repoName = parts[2] || "invalid";
        } catch (e) {
          repoOwner = "invalid";
          repoName = "invalid";
        }
      }

      gtag("event", "create_task", {
        event_category: "analyze",
        event_label: `${repoOwner}/${repoName}`,
      });
    }

    this.#resLayout.renderStatus(STATUS_INIT);

    client
      .createTask(formData)
      .then(handle)
      .catch(() => this.#resLayout.renderError("INTERNAL_ERROR"));
  }

  /**
   * @returns {Promise<void>}
   */
  getTaskStatus() {
    /**
     * @param {Partial<import("./client.js").TaskStatusResponse>} data
     * @returns {void}
     */
    const handleData = (data) => {
      this.#resLayout.renderStatus(data.task_status);

      if (data.task_done) {
        clearInterval(this.#fetchStatusInterval);

        if (data.task_error) {
          this.#form.disable = false;
          this.#hasAccess = false;
          this.#resLayout.renderError(data.task_error_message);
          return;
        }

        this.getTaskResult();
      }
    };

    const handleException = () => {
      this.#resLayout.renderError("INTERNAL_ERROR");
      clearInterval(this.#fetchStatusInterval);
    };

    return client.getTaskStatus(this.#taskID).then(handleData).catch(handleException);
  }

  /**
   * @returns {Promise<void>}
   */
  getTaskResult() {
    /**
     * @param {import("./client.js").TaskResult | string} data
     */
    const handle = (data) => {
      if (data.task_error) {
        this.#form.disable = false;
        this.#hasAccess = false;
        this.#resLayout.renderError(data.task_error_message);
        return;
      }

      this.#resData = data;
      this.getCSS();
      // $("head").append(
      //   $("<link>")
      //     .attr("href", "css/table.css")
      //     .attr("rel", "stylesheet"),
      // );

      this.#resLayout.renderResultJSON(this.#resData).then(() => {
        this.#hasAccess = true;
        this.#form.disable = false;
      });
    };

    return client
      .getTaskResult(this.#taskID, "application/json")
      .then(handle)
      .catch(() => this.#resLayout.renderError("INTERNAL_ERROR"));
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

    this.#elem.on("submit", (e) => this.props.onSubmit(e));
    $("#btn-option").on("click", () => $("#options").toggleClass("hidden"));

    $("#option-group-file").find(".input-option").attr("placeholder", this.randomPlaceholder("file"));

    $("#option-group-dir").find(".input-option").attr("placeholder", this.randomPlaceholder("dir"));

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

    $item.find(".input-text").attr("placeholder", this.randomPlaceholder(patternGroup)).val("");

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

  /**
   * @param {boolean} value
   */
  set disable(value) {
    $("button[type=submit]").prop("disabled", value);
  }
}

class ResponseLayout {
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
        ${svgIcons.error}
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
   * @description render result by html response
   * @deprecated
   */
  renderResutlHTML(result) {
    this.#error.addClass("hidden");

    return wait(1000)
      .then(() => {
        this.#elem.html(result);
        $("#form").removeClass("hidden");
      })
      .then(() => {
        $("html").animate({ scrollTop: $("#repo-table").offset().top }, 350);
      });
  }

  /**
   *	@param {import("./client.js").TaskResult} data
   *	@returns {Promise<void>}
   *	@description render result by json response
   */

  renderResultJSON(data) {
    this.#error.addClass("hidden");

    return wait()
      .then(() => {
        let metadata = $("<div>")
          .addClass("metadata")
          .append(`<p>Fetch Speed: <strong>${data.fetch_speed_str || "unknown"}</strong></p>`)
          .append(`<p>Analysis Speed: <strong>${data.analysis_speed_str || "unknown"}</strong></p>`)
          .append(`<p>Parallel Mode: <strong>${data.parallel_mode ? "yes" : "no"}</strong></p>`);

        let theadItems = [
          { text: "Language" },
          { text: "Files" },
          { text: "Lines" },
          { text: "Blank" },
          { text: "Comments" },
        ];
        /**
         *	@param {import("./client.js").TaskResult} data
         *	@param {string} field
         *	@param {"asc" | "desc"} order
         *	@returns {import("./client.js").TaskResult}
         */
        let sortFn = (data, field, order) => {
          field = field.toLowerCase();
          data.languages.sort((a, b) => {
            switch (order) {
              case "asc":
                if (typeof a[field] === "string") {
                  return a[field].localeCompare(b[field]);
                }
                return a[field] - b[field];
              case "desc":
                if (typeof a[field] === "string") {
                  return b[field].localeCompare(a[field]);
                }
                return b[field] - a[field];
            }
          });

          this.renderResultJSON({ ...data });
        };

        let sortInc = $("<button>").append(svgIcons.arrowUp).addClass("btn-sort");
        let sortDec = $("<button>").append(svgIcons.arrowDown).addClass("btn-sort");

        let thead = $("<thead>").append(
          $("<tr>").append(
            theadItems.map((item) => {
              return $("<th>").append(
                $("<div>")
                  .addClass("theader")
                  .append($("<span>").append(item.text))
                  .append(
                    $("<div>")
                      .addClass("th-buttons")
                      .append(sortInc.clone(false).on("click", () => sortFn(data, item.text, "asc")))
                      .append(sortDec.clone(false).on("click", () => sortFn(data, item.text, "desc"))),
                  ),
              );
            }),
          ),
        );

        let rows = data.languages.map(
          (lang) => `<tr>
					<td><img src="${lang.badge_url}"/></td>
					<td>${lang.files}</td>	
					<td>${lang.lines}</td>	
					<td>${lang.blank}</td>	
					<td>${lang.comments}</td>	
					</tr>`,
        );

        rows.push(`<tr>
					<td>TOTAL</td>
					<td>${data.total_files}</td>	
					<td>${data.total_lines}</td>	
					<td>${data.total_blank}</td>	
					<td>${data.total_comments}</td>	
					</tr>`);

        let tbody = $("<tbody>").append(rows);

        let table = $("<table>").addClass("repo-table").attr("id", "repo-table").append(thead, tbody);

        this.#elem.html($(`<div>`).addClass("main").append(metadata, table));
        $("#form").removeClass("hidden");
      })
      .then(() => $("html").animate({ scrollTop: $("#repo-table").offset().top }, 350));
  }
}

let mockResult = {
  repo_size_limit: 200,
  is_prod: false,
  parallel_mode: true,
  languages: [
    {
      name: "TypeScript",
      blank: 1297,
      comments: 392,
      lines: 11718,
      files: 409,
      badge_url: "https://img.shields.io/badge/TypeScript-3178C6?logo=typescript&logoColor=fff",
    },
    {
      name: "JavaScript",
      blank: 23,
      comments: 81,
      lines: 471,
      files: 11,
      badge_url: "https://img.shields.io/badge/JavaScript-F7DF1E?logo=javascript&logoColor=000",
    },
    {
      name: "JSON",
      blank: 0,
      comments: 0,
      lines: 284,
      files: 9,
      badge_url: "https://img.shields.io/badge/JSON-000?logo=json&logoColor=fff",
    },
    {
      name: "Other",
      blank: 25,
      comments: 0,
      lines: 188,
      files: 18,
      badge_url: "https://img.shields.io/badge/Other-000000?logo=github&logoColor=fff",
    },
    {
      name: "CSS",
      blank: 19,
      comments: 4,
      lines: 163,
      files: 1,
      badge_url: "https://img.shields.io/badge/CSS-1572B6?logo=css3&logoColor=fff",
    },
    {
      name: "Markdown",
      blank: 32,
      comments: 3,
      lines: 120,
      files: 2,
      badge_url: "https://img.shields.io/badge/Markdown-%23000000?logo=markdown&logoColor=white",
    },
    {
      name: "YAML",
      blank: 14,
      comments: 9,
      lines: 111,
      files: 2,
      badge_url: "https://img.shields.io/badge/YAML-CB171E?logo=yaml&logoColor=fff",
    },
    {
      name: "Bourne Shell",
      blank: 5,
      comments: 5,
      lines: 24,
      files: 2,
      badge_url: "https://img.shields.io/badge/Bash-4EAA25?logo=gnubash&logoColor=fff",
    },
    {
      name: "HTML",
      blank: 0,
      comments: 0,
      lines: 13,
      files: 1,
      badge_url: "https://img.shields.io/badge/HTML-%23E34F26?logo=html5&logoColor=white",
    },
  ],
  total_lines: 13092,
  total_files: 455,
  total_blank: 1415,
  total_comments: 494,
  fetch_speed: 1133821625,
  analysis_speed: 16120417,
  fetch_speed_str: "01.133 s",
  analysis_speed_str: "00.016 s",
  error: "",
};

$(() => {
  const app = new App();
  app.getCSS();
  app.resLayout.renderResultJSON(mockResult);
});
