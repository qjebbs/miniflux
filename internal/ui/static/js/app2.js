const policy = window.trustedTypes?.createPolicy('html', {
    createHTML: (input) => input
});

function findParent(element, selector) {
    for (; element && element !== document; element = element.parentNode) {
        if (element.classList.contains(selector)) {
            return element;
        }
    }
    return null;
}

// Handle set view action for feeds and categories pages.
function handleSetView(element) {
    if (!element) {
        return;
    }
    sendPOSTRequest(element.dataset.url, {
        view: element.dataset.value
    }).then((response) => {
        if (response.ok) location.reload();
    });
}

// Handle toggle NSFW action for pages.
function handleNSFW() {
    let element = document.querySelector("a[data-action=nsfw]");
    if (!element || !element.dataset.url) {
        return;
    }
    sendPOSTRequest(element.dataset.url).then((response) => {
        if (response.ok) location.reload();
    });
}

// Handle media cache from the list view and entry view.
function handleCache(element) {
    let currentEntry = findEntry(element);
    if (currentEntry) {
        toggleCache(document.querySelector(".entry"));
    }
}

function setCachedButtonState(buttonElement, newState) {
    buttonElement.dataset.value = newState;
    const iconType = newState === "cached" ? "uncache" : "cache";
    setIconAndLabelElement(buttonElement, iconType, buttonElement.dataset[newState === "cached" ? "labelUncached" : "labelCached"]);
}

// Send the Ajax request and change the icon when caching an entry.
function toggleCache(element) {
    const currentEntry = findEntry(element);
    if (!currentEntry) return;

    const buttonElement = currentEntry.querySelector(":is(a, button)[data-toggle-cache]");
    if (!buttonElement) return;

    setButtonToLoadingState(buttonElement);

    sendPOSTRequest(buttonElement.dataset.cacheUrl).then(() => {
        let currentStatus = buttonElement.dataset.value;
        let newStatus = currentStatus === "cached" ? "uncached" : "cached";
        let isCached = newStatus === "cached";

        setCachedButtonState(buttonElement, newStatus);

        if (isEntryView()) {
            const iconType = newStatus === "cached" ? "uncache" : "cache";
            showToastNotification(iconType, buttonElement.dataset[isCached ? "toastCached" : "toastUncached"]);
        }
    });
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
    let items = getVisibleElements(".items .item");
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
        imgs.forEach(img => {
            img.addEventListener("load", (e) => {
                if (layoutCallback) layoutCallback();
            })
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

// Function to handle lazy loading for a given container.
function initializeLazyLoad(container) {
    var lazyImages = [].slice.call(container.querySelectorAll("img.lazy"));

    if ("IntersectionObserver" in window) {
        let lazyImageObserver = new IntersectionObserver(function (entries, observer) {
            entries.forEach(function (entry) {
                if (entry.isIntersecting) {
                    let lazyImage = entry.target;
                    const originalSrc = lazyImage.dataset.src;
                    const fallbackSrc = lazyImage.dataset.fallbackSrc;

                    if (originalSrc) {
                        const img = new Image();
                        img.src = originalSrc;
                        img.onload = function () {
                            lazyImage.src = originalSrc;
                        };
                        img.onerror = function () {
                            // If the original image fails, use the fallback.
                            if (fallbackSrc) {
                                lazyImage.src = fallbackSrc;
                            }
                        };
                    }

                    lazyImage.classList.remove("lazy");
                    observer.unobserve(lazyImage);
                }
            });
        }, {
            rootMargin: '0px 0px 1000px 0px'
        });

        lazyImages.forEach(function (lazyImage) {
            lazyImageObserver.observe(lazyImage);
        });
    }
}

function initializeLazyLoadWithObserver() {
    // Initial load for the whole document.
    initializeLazyLoad(document);

    // Create a MutationObserver to watch for new nodes being added.
    const content = document.getElementById("content");
    if (content) {
        const mutationObserver = new MutationObserver(function (mutations) {
            mutations.forEach(function (mutation) {
                mutation.addedNodes.forEach(function (node) {
                    if (node.nodeType === 1) { // Check if it's an element.
                        initializeLazyLoad(node);
                    }
                });
            });
        });

        mutationObserver.observe(content, { childList: true, subtree: true });
    }
}

function initializeForkClickHandlers() {
    // Entry actions
    onClick(":is(a, button)[data-mark-above-read]", (event) => setEntriesAboveStatusRead(event.target));
    onClick(":is(a, button)[data-toggle-cache]", (event) => handleCache(event.target));
    onClick(":is(a, button)[data-set-read='true']", (event) => setEntryStatusRead(findEntry(event.target)), true);

    onClick(":is(a, button)[data-action=setView]", (event) => handleSetView(event.target));
    onClick(":is(a, button)[data-action=nsfw]", () => handleNSFW());
    onClick(":is(a, button)[data-action=historyGoBack]", () => history.back());

    let tabHandler = new TabHandler();
    tabHandler.addEventListener('.tabs.tabs-entry-edit', (header, content, i) => {
        let preview = document.querySelector('#preview-content');
        let editor = document.querySelector('#form-content');
        if (i == 0) {
            editor.value = preview.innerHTML;
        } else {
            preview.innerHTML = policy.createHTML(editor.value);
        }
    });
    onClick("button[data-action=submitEntry]", (event) => {
        let preview = document.querySelector('#preview-content');
        let editor = document.querySelector('#form-content');
        let previewParent = findParent(preview, "tab-content");
        if (previewParent.classList.contains('active')) {
            editor.value = preview.innerHTML;
        }
        document.querySelector("#entry-form").submit();
    });

}

initializeForkClickHandlers();
document.addEventListener("DOMContentLoaded", () => {
    initMasonryLayout();
    category_feeds_cascader();

    if (document.querySelector('.no-back-forward-cache')) {
        window.onpageshow = function (event) {
            if (event.persisted) {
                window.location.reload();
            }
        };
    }

    initializeLazyLoadWithObserver();
});