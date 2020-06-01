document.addEventListener("DOMContentLoaded", function () {
    handleSubmitButtons();
    initMasonryLayout();
    initTouchHandlers();

    if (!document.querySelector("body[data-disable-keyboard-shortcuts=true]")) {
        let keyboardHandler = new KeyboardHandler();
        keyboardHandler.on("g u", () => goToPage("unread"));
        keyboardHandler.on("g b", () => goToPage("starred"));
        keyboardHandler.on("g h", () => goToPage("history"));
        keyboardHandler.on("g f", () => goToFeedOrFeeds());
        keyboardHandler.on("g c", () => goToPage("categories"));
        keyboardHandler.on("g s", () => goToPage("settings"));
        keyboardHandler.on("ArrowLeft", () => goToPrevious());
        keyboardHandler.on("ArrowRight", () => goToNext());
        keyboardHandler.on("k", () => goToPrevious());
        keyboardHandler.on("p", () => goToPrevious());
        keyboardHandler.on("j", () => goToNext());
        keyboardHandler.on("n", () => goToNext());
        keyboardHandler.on("h", () => goToPage("previous"));
        keyboardHandler.on("l", () => goToPage("next"));
        keyboardHandler.on("o", () => openSelectedItem());
        keyboardHandler.on("v", () => openOriginalLink());
        keyboardHandler.on("V", () => openOriginalLink(true));
        keyboardHandler.on("c", () => openCommentLink());
        keyboardHandler.on("C", () => openCommentLink(true));
        keyboardHandler.on("m", () => handleEntryStatus());
        keyboardHandler.on("A", () => markPageAsRead());
        keyboardHandler.on("s", () => handleSaveEntry());
        keyboardHandler.on("d", () => handleFetchOriginalContent());
        keyboardHandler.on("f", () => handleBookmark());
        keyboardHandler.on("R", () => handleRefreshAllFeeds());
        keyboardHandler.on("?", () => showKeyboardShortcuts());
        keyboardHandler.on("#", () => unsubscribeFromFeed());
        keyboardHandler.on("/", (e) => setFocusToSearchInput(e));
        keyboardHandler.on("Escape", () => ModalHandler.close());
        keyboardHandler.on("P", () => forceProxyImages());
        keyboardHandler.on("N", () => handleNSFW());
        keyboardHandler.listen();
    }

    onClick("a[data-save-entry]", (event) => handleSaveEntry(event.target));
    onClick("a[data-toggle-bookmark]", (event) => handleBookmark(event.target));
    onClick("a[data-fetch-content-entry]", () => handleFetchOriginalContent());
    onClick("a[data-action=search]", (event) => setFocusToSearchInput(event));
    onClick("a[data-action=setView]", (event) => handleSetView(event.target));
    onClick("a[data-action=markPageAsRead]", () => handleConfirmationMessage(event.target, () => markPageAsRead()));
    onClick("a[data-toggle-status]", (event) => handleEntryStatus(event.target));
    onClick("a[data-action=nsfw]", () => handleNSFW());

    let tabHandler = new TabHandler();
    tabHandler.addEventListener('.tabs.tabs-entry-edit', EntryEditorHandler.switchHandler);

    onClick("a[data-toggle-cache]", (event) => handleCache(event.target));
    onClick("a[data-set-read]", (event) => setEntryStatusRead(findEntry(event.target)), true);
    onClick("a[data-action=showActionMenu]", (event) => ActionMenu.switch(event.target));
    onClick("button[data-action=submitEntry]", (event) => EntryEditorHandler.submitHandler(event));

    onClick("a[data-confirm]", (event) => handleConfirmationMessage(event.target, (url, redirectURL) => {
        let request = new RequestBuilder(url);

        request.withCallback(() => {
            if (redirectURL) {
                window.location.href = redirectURL;
            } else {
                window.location.reload();
            }
        });

        request.execute();
    }));

    if (document.documentElement.clientWidth < 600) {
        onClick(".logo", () => toggleMainMenu());
        onClick(".header nav li", (event) => onClickMainMenuListItem(event));
    }

    if (document.querySelector('.no-back-forward-cache')) {
        window.onpageshow = function (event) {
            if (event.persisted) {
                window.location.reload();
            }
        };
    }

    if ("serviceWorker" in navigator) {
        let scriptElement = document.getElementById("service-worker-script");
        if (scriptElement) {
            navigator.serviceWorker.register(scriptElement.src);
        }
    }

    window.addEventListener('beforeinstallprompt', (e) => {
        // Prevent Chrome 67 and earlier from automatically showing the prompt.
        e.preventDefault();

        let deferredPrompt = e;
        const promptHomeScreen = document.getElementById('prompt-home-screen');
        if (promptHomeScreen) {
            promptHomeScreen.style.display = "block";

            const btnAddToHomeScreen = document.getElementById('btn-add-to-home-screen');
            if (btnAddToHomeScreen) {
                btnAddToHomeScreen.addEventListener('click', (e) => {
                    e.preventDefault();
                    deferredPrompt.prompt();
                    deferredPrompt.userChoice.then(() => {
                        deferredPrompt = null;
                        promptHomeScreen.style.display = "none";
                    });
                });
            }
        }
    });
});
