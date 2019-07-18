// OnClick attaches a listener to the elements that match the selector.
function onClick(selector, callback, noPreventDefault) {
    let elements = document.querySelectorAll(selector);
    elements.forEach((element) => {
        element.onclick = (event) => {
            if (!noPreventDefault) {
                event.preventDefault();
            }

            callback(event);
        };
    });
}

// Show and hide the main menu on mobile devices.
function toggleMainMenu() {
    let menu = document.querySelector(".header nav ul");
    if (DomHelper.isVisible(menu)) {
        menu.style.display = "none";
    } else {
        menu.style.display = "block";
    }

    let searchElement = document.querySelector(".header .search");
    if (DomHelper.isVisible(searchElement)) {
        searchElement.style.display = "none";
    } else {
        searchElement.style.display = "block";
    }
}

// Handle click events for the main menu (<li> and <a>).
function onClickMainMenuListItem(event) {
    let element = event.target;

    if (element.tagName === "A") {
        window.location.href = element.getAttribute("href");
    } else {
        window.location.href = element.querySelector("a").getAttribute("href");
    }
}

// Change the button label when the page is loading.
function handleSubmitButtons() {
    let elements = document.querySelectorAll("form");
    elements.forEach((element) => {
        element.onsubmit = () => {
            let button = document.querySelector("button");

            if (button) {
                button.innerHTML = button.dataset.labelLoading;
                button.disabled = true;
            }
        };
    });
}

// Set cursor focus to the search input.
function setFocusToSearchInput(event) {
    event.preventDefault();
    event.stopPropagation();

    let toggleSwitchElement = document.querySelector(".search-toggle-switch");
    if (toggleSwitchElement) {
        toggleSwitchElement.style.display = "none";
    }

    let searchFormElement = document.querySelector(".search-form");
    if (searchFormElement) {
        searchFormElement.style.display = "block";
    }

    let searchInputElement = document.getElementById("search-input");
    if (searchInputElement) {
        searchInputElement.focus();
        searchInputElement.value = "";
    }
}

// Show modal dialog with the list of keyboard shortcuts.
function showKeyboardShortcuts() {
    let template = document.getElementById("keyboard-shortcuts");
    if (template !== null) {
        ModalHandler.open(template.content);
    }
}

// Mark as read visible items of the current page.
function markPageAsRead() {
    let items = DomHelper.getVisibleElements(".items .item");
    let entryIDs = [];

    items.forEach((element) => {
        element.classList.add("item-status-read");
        entryIDs.push(parseInt(element.dataset.id, 10));
    });

    if (entryIDs.length > 0) {
        updateEntriesStatus(entryIDs, "read", () => {
            // Make sure the Ajax request reach the server before we reload the page.

            let element = document.querySelector("a[data-action=markPageAsRead]");
            let showOnlyUnread = false;
            if (element) {
                showOnlyUnread = element.dataset.showOnlyUnread || false;
            }

            if (showOnlyUnread) {
                window.location.reload();
            } else {
                goToPage("next", true);
            }
        });
    }
}

// Handle entry status changes from the list view and entry view.
function handleEntryStatus() {
    if (isListView()) {
        let currentItem = document.querySelector(".current-item");
        if (currentItem !== null) {
            // The order is important here,
            // On the unread page, the read item will be hidden.
            goToNextListItem();
            toggleEntryStatus(currentItem);
        }
    } else {
        toggleEntryStatus(document.querySelector(".entry"));
    }
}

// Change the entry status to the opposite value.
function toggleEntryStatus(element) {
    let entryID = parseInt(element.dataset.id, 10);
    let link = element.querySelector("a[data-toggle-status]");

    let currentStatus = link.dataset.value;
    let newStatus = currentStatus === "read" ? "unread" : "read";

    updateEntriesStatus([entryID], newStatus);

    if (currentStatus === "read") {
        link.innerHTML = link.dataset.labelRead;
        link.dataset.value = "unread";
    } else {
        link.innerHTML = link.dataset.labelUnread;
        link.dataset.value = "read";
    }

    if (element.classList.contains("item-status-" + currentStatus)) {
        element.classList.remove("item-status-" + currentStatus);
        element.classList.add("item-status-" + newStatus);
    }
}

// Mark a single entry as read.
function markEntryAsRead(element) {
    if (element.classList.contains("item-status-unread")) {
        element.classList.remove("item-status-unread");
        element.classList.add("item-status-read");

        let entryID = parseInt(element.dataset.id, 10);
        updateEntriesStatus([entryID], "read");
    }
}

// Send the Ajax request to change entries statuses.
function updateEntriesStatus(entryIDs, status, callback) {
    let url = document.body.dataset.entriesStatusUrl;
    let request = new RequestBuilder(url);
    request.withBody({ entry_ids: entryIDs, status: status });
    request.withCallback(callback);
    request.execute();

    if (status === "read") {
        decrementUnreadCounter(1);
    } else {
        incrementUnreadCounter(1);
    }
}

// Handle save entry from list view and entry view.
function handleSaveEntry() {
    if (isListView()) {
        let currentItem = document.querySelector(".current-item");
        if (currentItem !== null) {
            saveEntry(currentItem.querySelector("a[data-save-entry]"));
        }
    } else {
        saveEntry(document.querySelector("a[data-save-entry]"));
    }
}

// Send the Ajax request to save an entry.
function saveEntry(element) {
    if (!element) {
        return;
    }

    if (element.dataset.completed) {
        return;
    }

    element.innerHTML = element.dataset.labelLoading;

    let request = new RequestBuilder(element.dataset.saveUrl);
    request.withCallback(() => {
        element.innerHTML = element.dataset.labelDone;
        element.dataset.completed = true;
    });
    request.execute();
}

// Handle bookmark from the list view and entry view.
function handleBookmark() {
    if (isListView()) {
        let currentItem = document.querySelector(".current-item");
        if (currentItem !== null) {
            toggleBookmark(currentItem);
        }
    } else {
        toggleBookmark(document.querySelector(".entry"));
    }
}

// Send the Ajax request and change the icon when bookmarking an entry.
function toggleBookmark(parentElement) {
    let element = parentElement.querySelector("a[data-toggle-bookmark]");
    if (!element) {
        return;
    }

    element.innerHTML = element.dataset.labelLoading;

    let request = new RequestBuilder(element.dataset.bookmarkUrl);
    request.withCallback(() => {
        if (element.dataset.value === "star") {
            element.innerHTML = element.dataset.labelStar;
            element.dataset.value = "unstar";
        } else {
            element.innerHTML = element.dataset.labelUnstar;
            element.dataset.value = "star";
        }
    });
    request.execute();
}

// Handle cache from the list view and entry view.
function handleCache() {
    if (isListView()) {
        let currentItem = document.querySelector(".current-item");
        if (currentItem !== null) {
            toggleCache(currentItem);
        }
    } else {
        toggleCache(document.querySelector(".entry"));
    }
}

// Send the Ajax request and change the icon when bookmarking an entry.
function toggleCache(parentElement) {
    let element = parentElement.querySelector("a[data-toggle-cache]");
    if (!element) {
        return;
    }

    element.innerHTML = element.dataset.labelLoading;

    let request = new RequestBuilder(element.dataset.cacheUrl);
    request.withCallback(() => {
        if (element.dataset.value === "cached") {
            element.innerHTML = element.dataset.labelCached;
            element.dataset.value = "uncached";
        } else {
            element.innerHTML = element.dataset.labelUncached;
            element.dataset.value = "cached";
        }
    });
    request.execute();
}

function setEntryStatusRead(element){
    let link = element.querySelector("a[data-set-read]");
    let sendRequest = !link.dataset.noRequest;
    if(sendRequest){
        let entryID = parseInt(element.dataset.id, 10);
        updateEntriesStatus([entryID], "read");
    }

    link = element.querySelector("a[data-toggle-status]");
    if (link && link.dataset.value === "unread") {
        link.innerHTML = link.dataset.labelUnread;
        link.dataset.value = "read";
    }

    if (element && element.classList.contains("item-status-unread")) {
        element.classList.remove("item-status-unread");
        element.classList.add("item-status-read");
        updateUnreadCounterValue
        decrementUnreadCounter(1);
    }
}

// Send the Ajax request to download the original web page.
function handleFetchOriginalContent() {
    if (isListView()) {
        return;
    }

    let element = document.querySelector("a[data-fetch-content-entry]");
    if (!element) {
        return;
    }

    if (element.dataset.completed) {
        return;
    }

    element.innerHTML = element.dataset.labelLoading;

    let request = new RequestBuilder(element.dataset.fetchContentUrl);
    request.withCallback((response) => {
        element.innerHTML = element.dataset.labelDone;
        element.dataset.completed = true;

        response.json().then((data) => {
            if (data.hasOwnProperty("content")) {
                document.querySelector(".entry-content").innerHTML = data.content;
            }
        });
    });
    request.execute();
}

function openOriginalLink() {
    let entryLink = document.querySelector(".entry h1 a");
    if (entryLink !== null) {
        DomHelper.openNewTab(entryLink.getAttribute("href"));
        return;
    }

    let currentItemOriginalLink = document.querySelector(".current-item a[data-original-link]");
    if (currentItemOriginalLink !== null) {
        DomHelper.openNewTab(currentItemOriginalLink.getAttribute("href"));

        // Move to the next item and if we are on the unread page mark this item as read.
        let currentItem = document.querySelector(".current-item");
        goToNextListItem();
        markEntryAsRead(currentItem);
    }
}

function openSelectedItem() {
    let currentItemLink = document.querySelector(".current-item .item-title a");
    if (currentItemLink !== null) {
        window.location.href = currentItemLink.getAttribute("href");
    }
}

function unsubscribeFromFeed() {
    let unsubscribeLinks = document.querySelectorAll("[data-action=remove-feed]");
    if (unsubscribeLinks.length === 1) {
        let unsubscribeLink = unsubscribeLinks[0];

        let request = new RequestBuilder(unsubscribeLink.dataset.url);
        request.withCallback(() => {
            if (unsubscribeLink.dataset.redirectUrl) {
                window.location.href = unsubscribeLink.dataset.redirectUrl;
            } else {
                window.location.reload();
            }
        });
        request.execute();
    }
}

/**
 * @param {string} page Page to redirect to.
 * @param {boolean} fallbackSelf Refresh actual page if the page is not found.
 */
function goToPage(page, fallbackSelf) {
    let element = document.querySelector("a[data-page=" + page + "]");

    if (element) {
        document.location.href = element.href;
    } else if (fallbackSelf) {
        window.location.reload();
    }
}

function goToPrevious() {
    if (isListView()) {
        goToPreviousListItem();
    } else {
        goToPage("previous");
    }
}

function goToNext() {
    if (isListView()) {
        goToNextListItem();
    } else {
        goToPage("next");
    }
}

function goToFeedOrFeeds() {
    if (isEntry()) {
        let feedAnchor = document.querySelector("span.entry-website a");
        if (feedAnchor !== null) {
            window.location.href = feedAnchor.href;
        }
    } else {
        goToPage('feeds');
    }
}

function goToPreviousListItem() {
    let items = DomHelper.getVisibleElements(".items .item");
    if (items.length === 0) {
        return;
    }

    if (document.querySelector(".current-item") === null) {
        items[0].classList.add("current-item");
        items[0].querySelector('.item-header a').focus();
        return;
    }

    for (let i = 0; i < items.length; i++) {
        if (items[i].classList.contains("current-item")) {
            items[i].classList.remove("current-item");

            if (i - 1 >= 0) {
                items[i - 1].classList.add("current-item");
                DomHelper.scrollPageTo(items[i - 1]);
                items[i - 1].querySelector('.item-header a').focus();
            }

            break;
        }
    }
}

function goToNextListItem() {
    let currentItem = document.querySelector(".current-item");
    let items = DomHelper.getVisibleElements(".items .item");
    if (items.length === 0) {
        return;
    }

    if (currentItem === null) {
        items[0].classList.add("current-item");
        items[0].querySelector('.item-header a').focus();
        return;
    }

    for (let i = 0; i < items.length; i++) {
        if (items[i].classList.contains("current-item")) {
            items[i].classList.remove("current-item");

            if (i + 1 < items.length) {
                items[i + 1].classList.add("current-item");
                DomHelper.scrollPageTo(items[i + 1]);
                items[i + 1].querySelector('.item-header a').focus();
            }

            break;
        }
    }
}

function decrementUnreadCounter(n) {
    updateUnreadCounterValue((current) => {
        return current - n;
    });
}

function incrementUnreadCounter(n) {
    updateUnreadCounterValue((current) => {
        return current + n;
    });
}

function updateUnreadCounterValue(callback) {
    let counterElements = document.querySelectorAll("span.unread-counter");
    counterElements.forEach((element) => {
        let oldValue = parseInt(element.textContent, 10);
        element.innerHTML = callback(oldValue);
    });

    if (window.location.href.endsWith('/unread')) {
        let oldValue = parseInt(document.title.split('(')[1], 10);
        let newValue = callback(oldValue);

        document.title = document.title.replace(
            /(.*?)\(\d+\)(.*?)/,
            function (match, prefix, suffix, offset, string) {
                return prefix + '(' + newValue + ')' + suffix;
            }
        );
    }
}

function isEntry() {
    return document.querySelector("section.entry") !== null;
}

function isListView() {
    return document.querySelector(".items") !== null;
}

function handleConfirmationMessage(linkElement, callback) {
    linkElement.style.display = "none";

    let containerElement = linkElement.parentNode;
    let questionElement = document.createElement("span");

    let yesElement = document.createElement("a");
    yesElement.href = "#";
    yesElement.appendChild(document.createTextNode(linkElement.dataset.labelYes));
    yesElement.onclick = (event) => {
        event.preventDefault();

        let loadingElement = document.createElement("span");
        loadingElement.className = "loading";
        loadingElement.appendChild(document.createTextNode(linkElement.dataset.labelLoading));

        questionElement.remove();
        containerElement.appendChild(loadingElement);

        callback(linkElement.dataset.url, linkElement.dataset.redirectUrl);
    };

    let noElement = document.createElement("a");
    noElement.href = "#";
    noElement.appendChild(document.createTextNode(linkElement.dataset.labelNo));
    noElement.onclick = (event) => {
        event.preventDefault();
        linkElement.style.display = "inline";
        questionElement.remove();
    };

    questionElement.className = "confirm";
    questionElement.appendChild(document.createTextNode(linkElement.dataset.labelQuestion + " "));
    questionElement.appendChild(yesElement);
    questionElement.appendChild(document.createTextNode(", "));
    questionElement.appendChild(noElement);

    containerElement.appendChild(questionElement);
}
