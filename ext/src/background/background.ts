import {
    sendPageData,
    sendResult,
} from '../modules/network';

const missingURLMsg = {"error": "Missing or invalid Hister server URL. Configure it in the addon popup."};
// TODO check source
function cjsMsgHandler(request, sender, sendResponse) {
    chrome.storage.local.get(['histerURL']).then(data => {
        let u = data['histerURL'] || "";
        if(!u) {
            chrome.tabs.sendMessage(sender.tab.id, missingURLMsg);
            return;
        }
        if(!u.endsWith('/')) {
            u += '/';
        }
        if(request.pageData) {
            sendPageData(u+"add", request.pageData).then((r) => sendResponse({"status": "ok", "status_code": r.status})).catch(err => sendResponse({"error": err.message}));
            return true;
        }
        if(request.resultData) {
            sendResult(u+"history", request.resultData).then((r) => sendResponse({"status": "ok", "status_code": r.status})).catch(err => sendResponse({"error": err.message}));
            return true;
        }
    }).catch(error => {
        chrome.tabs.sendMessage(sender.tab.id, missingURLMsg);
    });
    return true;
}

chrome.runtime.onMessage.addListener(cjsMsgHandler);
