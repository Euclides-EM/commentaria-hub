// ==UserScript==
// @name         Transkribus - Download Plain Text Transcription
// @namespace    https://transkribus.eu/
// @version      1.1
// @description  Download the current Transkribus page transcription as plain text
// @match        https://app.transkribus.org/sites/noscemus/doc/*
// @grant        GM_download
// @grant        GM_registerMenuCommand
// @grant        window.close
// @run-at       document-idle
// ==/UserScript==

(function () {
    'use strict';

    const timeout = setTimeout(() => {
        window.close();
    }, 10000);

    function getTranscription() {
        const container = document.querySelector('#textContainer');

        if (!container) {
            return null;
        }

        const lines = Array.from(
            container.querySelectorAll('[id^="text-"]')
        );

        if (!lines.length) {
            return null;
        }

        return lines
            .map(el => el.innerText.trim())
            .filter(Boolean)
            .join('\n');
    }

    function hasNoTranscription() {
        return Array.from(document.querySelectorAll('body *'))
            .some(el => el.children.length === 0 &&
                el.textContent.trim() === 'No transcription for this page');
    }

    function getPageNumber() {
        const match = document.body.innerText.match(/\bPage\s+(\d+)\b/i);

        if (!match) {
            throw new Error('Could not determine page number');
        }

        return match[1];
    }

    function getDocumentId() {
        // Try URL first.
        const urlMatch = location.href.match(
            /(?:document|documents|doc|docid)[/=](\d+)/i
        );

        if (urlMatch) {
            return urlMatch[1];
        }

        // Try links.
        for (const link of document.querySelectorAll('a[href]')) {
            const match = link.href.match(
                /(?:document|documents|doc|docid)[/=](\d+)/i
            );

            if (match) {
                return match[1];
            }
        }

        throw new Error('Could not determine document ID');
    }

    function download(text) {
        const page = getPageNumber();
        const docId = getDocumentId();

        const filename = `${docId}_${page}.txt`;

        const blob = new Blob(
            [text],
            { type: 'text/plain;charset=utf-8' }
        );

        const url = URL.createObjectURL(blob);

        GM_download({
            url,
            name: filename,
            saveAs: false,
            onload: () => URL.revokeObjectURL(url),
            onerror: (error) => {
                URL.revokeObjectURL(url);
                console.error(
                    '[Transkribus] Download failed:',
                    error
                );
            }
        });

        // Give the browser 1 second to start the download.
        setTimeout(() => {
            window.close();
        }, 1000);
    }

    function tryDownload() {
        // No transcription is a valid final state.
        if (hasNoTranscription()) {
            download('');
            return true;
        }

        // Transcription is available.
        const text = getTranscription();

        if (text !== null) {
            download(text);
            return true;
        }

        return false;
    }

    function waitForTranscription() {
        // Already available.
        if (tryDownload()) {
            return;
        }

        // Nuxt renders the editor asynchronously.
        const observer = new MutationObserver(() => {
            if (tryDownload()) {
                observer.disconnect();
            }
        });

        observer.observe(document.body, {
            childList: true,
            subtree: true
        });
    }

    waitForTranscription();
})();
