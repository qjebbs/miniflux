document.addEventListener("DOMContentLoaded", function () {
    handleSubmitButtons();

    let tabHandler = new TabHandler();
    tabHandler.addEventListener('.tabs.tabs-entry-edit', EntryEditorHandler.switchHandler);

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
        keyboardHandler.on("m", () => handleEntryStatus());
        keyboardHandler.on("A", () => markPageAsRead());
        keyboardHandler.on("s", () => handleSaveEntry());
        keyboardHandler.on("d", () => handleFetchOriginalContent());
        keyboardHandler.on("f", () => handleBookmark());
        keyboardHandler.on("?", () => showKeyboardShortcuts());
        keyboardHandler.on("#", () => unsubscribeFromFeed());
        keyboardHandler.on("/", (e) => setFocusToSearchInput(e));
        keyboardHandler.on("Escape", () => ModalHandler.close());
        keyboardHandler.listen();
    }

    let touchHandler = new TouchHandler();
    touchHandler.listen();

    onClick("a[data-save-entry]", () => handleSaveEntry());
    onClick("a[data-toggle-bookmark]", () => handleBookmark());
    onClick("a[data-toggle-cache]", () => handleCache());
    onClick("a[data-fetch-content-entry]", () => handleFetchOriginalContent());
    onClick("a[data-action=search]", (event) => setFocusToSearchInput(event));
    onClick("a[data-action=markPageAsRead]", () => handleConfirmationMessage(event.target, () => markPageAsRead()));
    onClick("a[data-action=historyGoBack]", () => { history.go(-1) });

    onClick("a[data-toggle-status]", (event) => {
        let currentItem = DomHelper.findParent(event.target, "entry");
        if (!currentItem) {
            currentItem = DomHelper.findParent(event.target, "item");
        }

        if (currentItem) {
            toggleEntryStatus(currentItem);
        }
    });

    onClick("a[data-set-read]", (event) => {
        let currentItem = DomHelper.findParent(event.target, "entry");
        if (!currentItem) {
            currentItem = DomHelper.findParent(event.target, "item");
        }
        if (currentItem) {
            setEntryStatusRead(currentItem)
        }

    }, true);

    onClick("a[data-action=showActionMenu]", (event) => {
        let currentItem = DomHelper.findParent(event.target, "entry");
        if (!currentItem) {
            currentItem = DomHelper.findParent(event.target, "item");
        }
        if (currentItem) {
            new ActionMenuHandler(currentItem).show();
        }
    })

    onClick("button[data-action=submitEntry]", (event) => {
        EntryEditorHandler.submitHandler(event);
    });

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

    if ("serviceWorker" in navigator) {
        let scriptElement = document.getElementById("service-worker-script");
        if (scriptElement) {
            navigator.serviceWorker.register(scriptElement.src);
        }
    }
});

window.onload = function () {
    // masonry has to wait for all resources loaded to get the right layout
    let msnryElement = document.querySelector('.masonry');
    if (msnryElement) {
        var msnry = new Masonry(msnryElement, {
            itemSelector: '.item',
            columnWidth: '.item-sizer',
            gutter: 10
        })
        let callback = (instance, image) => {
            if (image && image.img && !image.img.dataset.src && !image.isLoaded) {
                let thumbnail = DomHelper.findParent(image.img, "thumbnail");
                if (thumbnail) thumbnail.parentNode.removeChild(thumbnail);
            }
            msnry.layout();
        }
        imagesLoaded('.masonry .item').on('progress', callback);
        LazyloadHandler.add(".item", 'progress', callback);
    }
};