let ws;
let input = document.getElementById("search");
let results = document.getElementById("results");
let emptyImg = "data:image/gif;base64,R0lGODlhAQABAAAAACH5BAEKAAEALAAAAAABAAEAAAICTAEAOw==";
let urlState = {};
let templates = {};
for(let el of document.querySelectorAll("template")) {
    let id = el.getAttribute("id")
    templates[id] = el;
}

function createTemplate(name, fns) {
    let el = document.importNode(templates[name].content, true)
    for(let k in fns) {
        fns[k](el.querySelector(k));
    }
    return el;
}

function connect() {
    ws = new WebSocket(document.querySelector("#ws-url").value);

    ws.onopen = function() {
        const urlParams = new URLSearchParams(window.location.search);
		const query = urlParams.get('q');
        if(query) {
            sendQuery(query);
            input.value = query;
        }
    };

    ws.onmessage = renderResults;

    ws.onclose = function() {
        console.log("WebSocket connection closed, retrying...");
        setTimeout(connect, 1000); // Reconnect after 1 second
    };

    ws.onerror = function(error) {
        console.error("WebSocket error:", error);
    };
}

function sendQuery(q) {
    let message = {"text": q, "highlight": "HTML"};
    ws.send(JSON.stringify(message));
}

function updateURL() {
    if(input.value) {
        history.replaceState(urlState, "", `${window.location.pathname}?q=${encodeURIComponent(input.value)}`);
        return;
    }
    history.replaceState(urlState, "", `${window.location.pathname}`);
}

function renderResults(event) {
    const res = JSON.parse(event.data);
    const d = res.documents;
    if(!d || !d.length) {
        if(!input.value) {
            results.replaceChildren(createTemplate("tips", {}));
            return
        }
        let u = getSearchUrl(input.value)
        let n = createTemplate("result", {
            "a": (e) => { e.setAttribute("href", u); e.innerHTML = "No results found - open query in web search engine"; e.classList.add("error"); },
            ".result-url": (e) => { e.textContent = u; },
        });
        results.replaceChildren(n);
        return;
    }
    let resultElements = [];
    highlightIdx = 0;
    resultElements.push(createResultsHeader(res));
    if(res.history && res.history.length) {
        for(let r of res.history) {
            resultElements.push(createPriorityResult(r))
        }
    }
    for(let r of d) {
        resultElements.push(createResult(r));
    }
    results.replaceChildren(...resultElements);
};

input.addEventListener("input", () => {
    updateURL();
    sendQuery(input.value);
});

function getSearchUrl(query) {
    return document.querySelector("#search-url").value.replace("{query}", escape(query));
}

function openUrl(u, newWindow) {
    if(newWindow) {
        window.open(u, '_blank');
        window.focus();
        return;
    }
    window.location.href = u;
}

function init() {
    results.replaceChildren(createTemplate("tips", {}));
    connect();
}

function openResult(e, newWindow) {
    if(e.preventDefault) {
        e.preventDefault();
    }
    let url = e.target.getAttribute("href");
    let title = e.target.innerText;
    saveHistoryItem(url, title, input.value).then((r) => {
		openUrl(url, newWindow);
	});
    return false;
}

function createResultsHeader(res) {
    const d = res.documents;
    const header = createTemplate("results-header", {
        ".duration": (e) => e.innerText = res.search_duration || "",
        ".results-num": (e) => e.innerText = res.total || d.length,
    });
    if(res.query && res.query.text != input.value) {
        header.querySelector(".expanded-query").innerHTML = `Expanded query: <code>"${res.query.text}"</code>`;
    }
    return header;
}

function createPriorityResult(r) {
    let rn = createTemplate("result", {
        "a": (e) => {
            e.setAttribute("href", r.url);
            e.innerHTML = r.title || "*title*";
            // TODO handle middleclick (auxclick handler)
            e.addEventListener("click", openResult);
            e.classList.add("success");
        },
        ".result-url": (e) => { e.textContent = r.url; },
    });
    return rn;
}

function createResult(r) {
    let rn = createTemplate("result", {
        "a": (e) => {
            e.setAttribute("href", r.url);
            e.innerHTML = r.title || "*title*";
            // TODO handle middleclick (auxclick handler)
            e.addEventListener("click", openResult);
        },
        "img": (e) => { e.setAttribute("src", r.favicon || emptyImg); },
        ".result-url": (e) => { e.textContent = r.url; },
        ".action-button": (e) => { e.addEventListener("click", (ev) => toggleActions(ev, e.closest(".result"))) },
        "p": (e) => { e.innerHTML = r.text || ""; },
    });
    return rn;
}

function toggleActions(ev, res) {
    let a = res.querySelector(".actions")
    if(a) {
        closeActions(a);
        return;
    }
    a = createTemplate("result-actions", {
        ".save": (e) => e.addEventListener("click", () => savePriorityResult(e)),
        ".close": (e) => e.addEventListener("click", () => closeActions(e)),
    });
    for(let e of a.children) {
        e.style.animation = "fade-in 0.5s";
    }
    res.appendChild(a);
}

function closeActions(e) {
    e.closest(".actions").remove();
}

function savePriorityResult(e) {
    let result = e.closest(".result");
    let link = result.querySelector(".result-title a");
    let url = link.getAttribute("href");
    let title = link.innerText;
    let query = result.querySelector(".action-query").value;
    if(!query) {
        return;
    }
    saveHistoryItem(url, title, query).then((r) => {
        result.querySelector(".actions").appendChild(createTemplate("success", {
            ".message": (e) => e.innerText = "Priority result added.",
        }));
    });
}

function saveHistoryItem(url, title, query) {
	return fetch("/history", {
		method: "POST",
		body: JSON.stringify({"url": url, "title": title, "query": query}),
		headers: {"Content-type": "application/json; charset=UTF-8"},
	})
}

let highlightIdx = 0;
window.addEventListener("keydown", function(e) {
    if(e.key == "/") {
        if(document.activeElement != input) {
            e.preventDefault();
            input.focus();
        }
    }
    if(e.key == "Enter") {
        e.preventDefault();
        let res = document.querySelectorAll(".result a")[highlightIdx];
        let newWindow = e.ctrlKey ? true : false;
        openResult({'target': res}, newWindow);
    }
    if(e.ctrlKey && (e.key == "j" || e.key == "k")) {
          e.preventDefault();
          if(results.children.length > highlightIdx) {
              results.children[highlightIdx].classList.remove("highlight");
          }
          highlightIdx = (highlightIdx+(e.key=="j"?1:-1)+results.children.length) % results.children.length;
          results.children[highlightIdx].classList.add("highlight");
    }
    if(e.ctrlKey && e.key == "o") {
        e.preventDefault();
        openUrl(getSearchUrl(input.value));
    }
});

init();
