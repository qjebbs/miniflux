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
            let button = element.querySelector("button");

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
        let container = ModalHandler.open(template.content);
        let modal = container.querySelector(".modal");
        if (modal === null) return;
        setTimeout(() => {
            modal.classList.add("fade");
        }, 100)
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
                window.scrollTo(0, 0);
                window.location.reload();
            } else {
                goToPage("next", true);
            }
        });
    }
}

// Handle entry status changes from the list view and entry view.
function handleEntryStatus(element) {
    let toasting = !element;
    let currentEntry = findEntry(element);
    if (currentEntry) {
        toggleEntryStatus(currentEntry, toasting);
        if (isListView() && currentEntry.classList.contains('current-item')) {
            goToNextListItem();
        }
    }
}

// Change the entry status to the opposite value.
function toggleEntryStatus(element, toasting) {
    let entryID = parseInt(element.dataset.id, 10);
    let link = element.querySelector("a[data-toggle-status]");

    let currentStatus = link.dataset.value;
    let newStatus = currentStatus === "read" ? "unread" : "read";

    updateEntriesStatus([entryID], newStatus);

    let icon, label;

    if (currentStatus === "read") {
        icon = document.querySelector("template#icon_read");
        label = link.dataset.labelRead;
        if (toasting) {
            toast(link.dataset.toastUnread);
        }
    } else {
        icon = document.querySelector("template#icon_unread");
        label = link.dataset.labelUnread;
        if (toasting) {
            toast(link.dataset.toastRead);
        }
    }

    link.innerHTML = icon.innerHTML + '<span class="icon-label">' + label + '</span>';
    link.dataset.value = newStatus;

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

// Send the Ajax request to refresh all feeds in the background
function handleRefreshAllFeeds() {
    let url = document.body.dataset.refreshAllFeedsUrl;
    let request = new RequestBuilder(url);

    request.withCallback(() => {
        window.location.reload();
    });

    request.withHttpMethod("GET");
    request.execute();
}

// Send the Ajax request to change entries statuses.
function updateEntriesStatus(entryIDs, status, callback) {
    let url = document.body.dataset.entriesStatusUrl;
    let request = new RequestBuilder(url);
    request.withBody({entry_ids: entryIDs, status: status});
    request.withCallback(callback);
    request.execute();

    if (status === "read") {
        decrementUnreadCounter(1);
    } else {
        incrementUnreadCounter(1);
    }
}

// Handle save entry from list view and entry view.
function handleSaveEntry(element) {
    let toasting = !element;
    let currentEntry = findEntry(element);
    if (currentEntry) {
        saveEntry(currentEntry.querySelector("a[data-save-entry]"), toasting);
    }
}

// Handle set view action for feeds and categories pages.
function handleSetView(element) {
    if (!element) {
        return;
    }
    let request = new RequestBuilder(element.dataset.url);
    request.withForm({
        view: element.dataset.value
    });
    request.withCallback((response) => {
        if (response.ok) location.reload();
    });
    request.execute();
}

// Handle toggle NSFW action for pages.
function handleNSFW() {
    let element = document.querySelector("a[data-action=nsfw]");
    if (!element || !element.dataset.url) {
        return;
    }
    let request = new RequestBuilder(element.dataset.url);
    request.withCallback((response) => {
        if (response.ok) location.reload();
    });
    request.execute();
}

// Send the Ajax request to save an entry.
function saveEntry(element, toasting) {
    if (!element) {
        return;
    }

    if (element.dataset.completed) {
        return;
    }

    let previousInnerHTML = element.innerHTML;
    element.innerHTML = '<span class="icon-label">' + element.dataset.labelLoading + '</span>';

    let request = new RequestBuilder(element.dataset.saveUrl);
    request.withCallback(() => {
        element.innerHTML = previousInnerHTML;
        element.dataset.completed = true;
        if (toasting) {
            toast(element.dataset.toastDone);
        }
    });
    request.execute();
}

// Handle bookmark from the list view and entry view.
function handleBookmark(element) {
    let toasting = !element;
    let currentEntry = findEntry(element);
    if (currentEntry) {
        toggleBookmark(currentEntry, toasting);
    }
}

// Send the Ajax request and change the icon when bookmarking an entry.
function toggleBookmark(parentElement, toasting) {
    let element = parentElement.querySelector("a[data-toggle-bookmark]");
    if (!element) {
        return;
    }

    element.innerHTML = '<span class="icon-label">' + element.dataset.labelLoading + '</span>';

    let request = new RequestBuilder(element.dataset.bookmarkUrl);
    request.withCallback(() => {

        let currentStarStatus = element.dataset.value;
        let newStarStatus = currentStarStatus === "star" ? "unstar" : "star";

        let icon, label;

        if (currentStarStatus === "star") {
            icon = document.querySelector("template#icon_star");
            label = element.dataset.labelStar;
            if (toasting) {
                toast(element.dataset.toastUnstar);
            }
        } else {
            icon = document.querySelector("template#icon_unstar");
            label = element.dataset.labelUnstar;
            if (toasting) {
                toast(element.dataset.toastStar);
            }
        }

        element.innerHTML = icon.innerHTML + '<span class="icon-label">' + label + '</span>';
        element.dataset.value = newStarStatus;
    });
    request.execute();
}

// Handle media cache from the list view and entry view.
function handleCache(element) {
    let currentEntry = findEntry(element);
    if (currentEntry) {
        toggleCache(document.querySelector(".entry"));
    }
}

// Send the Ajax request and change the icon when caching an entry.
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

function setEntryStatusRead(element) {
    if (!element) return;
    let link = element.querySelector("a[data-set-read]");
    let sendRequest = !link.dataset.noRequest;
    if (sendRequest) {
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

function setEntriesAboveStatusRead(element) {
    let currentItem = findEntry(element);
    let items = DomHelper.getVisibleElements(".items .item");
    if (!currentItem || items.length === 0) {
        return;
    }
    let targetItems = [];
    let entryIds = [];
    for (let i = 0; i < items.length; i++) {
        if (items[i] == currentItem) {
            break;
        }
        targetItems.push(items[i]);
        entryIds.push(parseInt(items[i].dataset.id, 10));
    }
    updateEntriesStatus(entryIds, "read", () => {
        targetItems.map(item => {
            let link = item.querySelector("a[data-toggle-status]");
            if (link && link.dataset.value === "unread") {
                link.innerHTML = link.dataset.labelUnread;
                link.dataset.value = "read";
            }
            if (item && item.classList.contains("item-status-unread")) {
                item.classList.remove("item-status-unread");
                item.classList.add("item-status-read");
                decrementUnreadCounter(1);
            }
        });
    });
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

    let previousInnerHTML = element.innerHTML;
    element.innerHTML = '<span class="icon-label">' + element.dataset.labelLoading + '</span>';

    let request = new RequestBuilder(element.dataset.fetchContentUrl);
    request.withCallback((response) => {
        element.innerHTML = previousInnerHTML;

        response.json().then((data) => {
            if (data.hasOwnProperty("content")) {
                document.querySelector(".entry-content").innerHTML = data.content;
            }
        });
    });
    request.execute();
}

function openOriginalLink(openLinkInCurrentTab) {
    let entryLink = document.querySelector(".entry h1 a");
    if (entryLink !== null) {
        if (openLinkInCurrentTab) {
            window.location.href = entryLink.getAttribute("href");
        } else {
            DomHelper.openNewTab(entryLink.getAttribute("href"));
        }
        return;
    }

    let currentItemOriginalLink = document.querySelector(".current-item a[data-original-link]");
    if (currentItemOriginalLink !== null) {
        DomHelper.openNewTab(currentItemOriginalLink.getAttribute("href"));

        let currentItem = document.querySelector(".current-item");
        // If we are not on the list of starred items, move to the next item
        if (document.location.href != document.querySelector('a[data-page=starred]').href) {
            goToNextListItem();
        }
        markEntryAsRead(currentItem);
    }
}

function openCommentLink(openLinkInCurrentTab) {
    if (!isListView()) {
        let entryLink = document.querySelector("a[data-comments-link]");
        if (entryLink !== null) {
            if (openLinkInCurrentTab) {
                window.location.href = entryLink.getAttribute("href");
            } else {
                DomHelper.openNewTab(entryLink.getAttribute("href"));
            }
            return;
        }
    } else {
        let currentItemCommentsLink = document.querySelector(".current-item a[data-comments-link]");
        if (currentItemCommentsLink !== null) {
            DomHelper.openNewTab(currentItemCommentsLink.getAttribute("href"));
        }
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
        goToFeed();
    } else {
        goToPage('feeds');
    }
}

function goToFeed() {
    if (isEntry()) {
        let feedAnchor = document.querySelector("span.entry-website a");
        if (feedAnchor !== null) {
            window.location.href = feedAnchor.href;
        }
    } else {
        let currentItemFeed = document.querySelector(".current-item a[data-feed-link]");
        if (currentItemFeed !== null) {
            window.location.href = currentItemFeed.getAttribute("href");
        }
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

            let nextItem;
            if (i - 1 >= 0) {
                nextItem = items[i - 1];
            } else {
                nextItem = items[items.length - 1];
            }

            nextItem.classList.add("current-item");
            DomHelper.scrollPageTo(nextItem);
            nextItem.querySelector('.item-header a').focus();

            break;
        }
    }
}

function goToNextListItem() {
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

            let nextItem;
            if (i + 1 < items.length) {
                nextItem = items[i + 1];
            } else {
                nextItem = items[0];
            }

            nextItem.classList.add("current-item");
            DomHelper.scrollPageTo(nextItem);
            nextItem.querySelector('.item-header a').focus();

            break;
        }
    }
}

function scrollToCurrentItem() {
    let currentItem = document.querySelector(".current-item");
    if (currentItem !== null) {
        DomHelper.scrollPageTo(currentItem, true);
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

function findEntry(element) {
    if (isListView()) {
        if (element) {
            return DomHelper.findParent(element, "item");
        } else {
            return document.querySelector(".current-item");
        }
    } else {
        return document.querySelector(".entry");
    }
}

function handleConfirmationMessage(linkElement, callback) {
    if (linkElement.tagName != 'A') {
        linkElement = linkElement.parentNode;
    }

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

function initMasonryLayout() {
    let layoutCallback;
    let msnryElement = document.querySelector('.masonry');
    if (msnryElement) {
        let msnry = new Masonry(msnryElement, {
            itemSelector: '.item',
            columnWidth: '.item-sizer',
            gutter: 10
        })
        layoutCallback = throttle(() => msnry.layout(), 500, 1000);
        // initialize layout
        // important for layout of masonry view without images. e.g.: statistics page.
        layoutCallback();
    }
    let imgs = document.querySelectorAll(".masonry img");
    if (layoutCallback && imgs.length) {
        LazyloadHandler.add(".item", 'progress', layoutCallback);
        imgs.forEach(img => {
            img.addEventListener("error", (e) => {
                if (img && img.src == location.href) {
                    // should ignore no src error
                    // console.log("no src error");
                    return;
                }
                if (img) {
                    img.src = addProxyParam(img.src);
                    img = undefined;
                } else {
                    e.target.parentNode.removeChild(e.target);
                    layoutCallback();
                }
            })
        });
        return;
    }
    // try force proxy when failed, for entry imgages.
    imgs = document.querySelectorAll(".entry-content img");
    imgs.forEach(img => {
        img.addEventListener("error", (e) => {
            if (img) {
                img.src = addProxyParam(img.src);
                img = undefined;
            }
        })
    });
}

function forceProxyImages() {
    let imgs = document.querySelectorAll(".entry-content img");
    imgs.forEach(img => img.src = addProxyParam(img.src));
}

function addProxyParam(url) {
    let parts = url.split('?');
    let params = parts[1] ? parts[1].split('&') : [];
    for (let i = 0; i < params.length; i++) {
        if (params[i].toLowerCase().startsWith("proxy=")) {
            params.splice(i, 1);
        }
    }
    params.push("proxy=force");
    return parts[0] + '?' + params.join('&');
}

function toast(msg) {
    if (!msg) return;
    document.querySelector('.toast-wrap .toast-msg').innerHTML = msg;
    let toastWrapper = document.querySelector('.toast-wrap');
    toastWrapper.classList.remove('toastAnimate');
    setTimeout(function () {
        toastWrapper.classList.add('toastAnimate');
    }, 100);
}

function category_feeds_cascader() {
    let cata = document.querySelector('#form-category') // as HTMLSelectElement;
    let feed = document.querySelector('#form-feed') // as HTMLSelectElement;
    if (!cata || !feed) return;
    let span = document.createElement('span');
    feed.appendChild(span)
    cata.addEventListener("change", e => {
        // hide all options
        while (feed.options.length) {
            span.appendChild(feed.options[0])
        }
        for (let option of feed.querySelectorAll("span>option")) {
            if (!cata.value || cata.value == option.dataset.category) {
                feed.appendChild(option)
            }
        }
        return true;
    })
}

function throttle(fn, delay, atleast) {
    var timeout = null,
        startTime = new Date();
    return function (...args) {
        var curTime = new Date();
        clearTimeout(timeout);
        if (curTime - startTime >= atleast) {
            fn(...args);
            startTime = curTime;
        } else {
            timeout = setTimeout(() => fn(...args), delay);
        }
    }
}