"use strict";

$(function () {
    let userInput = $('#userInput');
    let saveEditButton = $('#saveEditButton');
    // Returns a function, that, as long as it continues to be invoked, will not
    // be triggered. The function will be called after it stops being called for
    // N milliseconds. If `immediate` is passed, trigger the function on the
    // leading edge, instead of the trailing.
    function debounce(func, wait, immediate) {
        var timeout;
        return function () {
            saveEditButton.removeClass()
            saveEditButton.text("Editing");
            var context = this,
                args = arguments;
            var later = function () {
                timeout = null;
                if (!immediate) func.apply(context, args);
            };
            var callNow = immediate && !timeout;
            clearTimeout(timeout);
            timeout = setTimeout(later, wait);
            if (callNow) func.apply(context, args);
        };
    }

    // This will apply the debounce effect on the keyup event
    // And it only fires 500ms or half a second after the user stopped typing
    var prevText = userInput.val();
    console.log("debounce: " + window.notes.debounceMS)
    userInput.on('keyup', debounce(function () {
        if (prevText === userInput.val()) {
            return // no changes
        }
        prevText = userInput.val();
        saveEditButton.removeClass()
        saveEditButton.text("Saving")
        upload();
    }, window.notes.debounceMS));

    var latestUpload = null, needAnother = false;

    function upload() {
        // Prevent concurrent uploads
        if (latestUpload != null) {
            needAnother = true;
            return
        }
        latestUpload = $.ajax({
            type: 'POST',
            url: '/update',
            data: JSON.stringify({
                new_text: userInput.val(),
                page: window.notes.pageName,
                fetched_at: window.lastFetch,
            }),
            success: function (data) {
                latestUpload = null;
                saveEditButton.removeClass()
                if (data.success === true) {
                    $("#rendered").html(data.rendered);
                    retagContent();
                    saveEditButton.addClass("success");
                    window.lastFetch = data.unix_time;
                    if (needAnother) {
                        upload();
                    }
                } else {
                    saveEditButton.addClass("failure");
                }
                saveEditButton.text(data.message);
                needAnother = false;
            },
            error: function (xhr, error) {
                latestUpload = null;
                needAnother = false;
                saveEditButton.removeClass()
                saveEditButton.addClass("failure");
                saveEditButton.text(error);
            },
            contentType: "application/json",
            dataType: 'json'
        });
    }

    retagContent();
});

function retagContent() {
    // Render checkmarks
    $("li:has(input)").addClass("checkmark done");
    $("ul:has(input)").addClass("checkmark");
    $("li.checkmark:has(input:not(:checked))").removeClass("done");
    $("input").click(function () {
        return false;
    });

    // Fix page title
    let h1 = $("h1");
    if (h1.length > 0) {
        document.title = h1.first().text();
    }

    // Center tables
    $("table").wrap("<div class='table-wrapper'></div>");

    // Re-render LaTeX equations
    if (window.MathJax.typeset !== undefined) {
        window.MathJax.texReset();
        window.MathJax.typesetClear();
        window.MathJax.typeset();
    }

    // Reapply syntax highlight
    hljs.initHighlighting.called = false;
    hljs.initHighlighting();

    // Re-render Mermaid diagrams
    mermaid.init();
}

