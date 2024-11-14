const ACTION_STATUS = 0;
const ACTION_RESULT = 1;

/**
 * @typedef {Object} Language
 * @property {string} name
 * @property {number} blank
 * @property {number} comments
 * @property {number} lines
 * @property {number} files
 * @property {string} badge_url
 */

/**
 * @typedef {Object} TaskResult
 * @property {number} repo_size_limit
 * @property {boolean} parallel_mode
 * @property {Language[]} languages
 * @property {number} total_lines
 * @property {number} total_files
 * @property {number} total_blank
 * @property {number} total_comments
 * @property {number} fetch_speed
 * @property {number} analysis_speed
 * @property {string} fetch_speed_str
 * @property {string} analysis_speed_str
 * @property {string} error
 */

/**
 *  POST "/"
 *	@typedef {Object} TaskInitResponse
 *	@property {string} id
 *	@property {boolean} error
 *	@property {string} error_message
 */

/**
 * GET "/task/:id/:action"
 * @typedef {Object} TaskStatusResponse
 * @property {number} task_status
 * @property {boolean} task_done
 * @property {boolean} task_error
 * @property {string} task_error_message
 * @property {TaskResult}  task_result
 */



class Client {
  /**
   *  @param {FormData} formData
   *  @returns {Promise<TaskInitResponse | string>}
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
   * @returns {Promise<TaskStatusResponse>}
   */
  getTaskStatus(taskID) {
    return fetch(`/api/task/${taskID}/${ACTION_STATUS}`, {
      method: "GET",
      headers: {
        Accept: "application/json",
      }
    }).then((response) => response.json());
  }

  /**
   *	@param {taskID} string
   *	@param {"application/json" | "text/html"} acceptType
   *	@returns {Promise<string | TaskResult>}
   */
  getTaskResult(taskID, acceptType) {
    return fetch(`/api/task/${taskID}/${ACTION_RESULT}`, {
      method: "GET",
      headers: {
        Accept: acceptType,
      }
    }).then((response) => acceptType.includes("text/html") ? response.text() : response.json());
  }
}

export const client = new Client();

