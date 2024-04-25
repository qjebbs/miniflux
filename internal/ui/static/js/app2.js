// Handle set view action for feeds and categories pages.
function handleSetView(element) {
    if (!element) {
        return;
    }
    let request = new RequestBuilder(element.dataset.url);
    request.withBody({
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

// Handle media cache from the list view and entry view.
function handleCache(element) {
    let currentEntry = findEntry(element);
    if (currentEntry) {
        toggleCache(document.querySelector(".entry"));
    }
}

// Send the Ajax request and change the icon when caching an entry.
function toggleCache(element) {
    let link = element.querySelector("a[data-toggle-cache]");
    if (!link) {
        return;
    }

    link.innerHTML = link.dataset.labelLoading;

    let request = new RequestBuilder(link.dataset.cacheUrl);
    request.withCallback(() => {
        let currentStatus = link.dataset.value;
        let newStatus = currentStatus === "cached" ? "uncached" : "cached";

        let iconElement, label;

        if (currentStatus === "cached") {
            iconElement = document.querySelector("template#icon-cache");
            label = link.dataset.labelCached;
        } else {
            iconElement = document.querySelector("template#icon-uncache");
            label = link.dataset.labelUncached;
        }

        link.innerHTML = iconElement.innerHTML + '<span class="icon-label">' + label + '</span>';
        link.dataset.value = newStatus;
    });
    request.execute();
}

function setEntryStatusRead(element) {
    if (!element || !element.classList.contains("item-status-unread")) {
        return;
    }
    let link = element.querySelector("a[data-set-read]");
    let sendRequest = !link.dataset.noRequest;
    if (!sendRequest) {
        handleEntryStatus("next", element, true);
        updateUnreadCounterValue(n => n - 1);
        return;
    }
    toggleEntryStatus(element, false);
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
    if (entryIds.length === 0) {
        return;
    }
    updateEntriesStatus(entryIds, "read", () => {
        targetItems.map(element => {
            handleEntryStatus("next", element, true);
        });
    });
}

// https://masonry.desandro.com
function initMasonryLayout() {
    let layoutCallback;
    let msnryElement = document.querySelector('.masonry');
    if (msnryElement) {
        let msnry = new Masonry(msnryElement, {
            itemSelector: '.item',
            columnWidth: '.item-sizer',
            gutter: 10,
            horizontalOrder: false,
            transitionDuration: '0.2s'
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
                if (layoutCallback) layoutCallback();
            })
        });
        return;
    }
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