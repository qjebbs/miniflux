class EntryHandler {
    static updateEntriesStatus(entryIDs, status, callback) {
        let url = document.body.dataset.entriesStatusUrl;
        let request = new RequestBuilder(url);
        request.withBody({entry_ids: entryIDs, status: status});
        request.withCallback(callback);
        request.execute();
    }

    static toggleEntryStatus(element) {
        let entryID = parseInt(element.dataset.id, 10);
        let link = element.querySelector("a[data-toggle-status]");

        let currentStatus = link.dataset.value;
        let newStatus = currentStatus === "read" ? "unread" : "read";

        this.updateEntriesStatus([entryID], newStatus);

        if (currentStatus === "read") {
            link.innerHTML = link.dataset.labelRead;
            link.dataset.value = "unread";
            UnreadCounterHandler.increment(1);
        } else {
            link.innerHTML = link.dataset.labelUnread;
            link.dataset.value = "read";
            UnreadCounterHandler.decrement(1);
        }

        if (element.classList.contains("item-status-" + currentStatus)) {
            element.classList.remove("item-status-" + currentStatus);
            element.classList.add("item-status-" + newStatus);
        }
    }

    static setEntryStatusRead(element){
        let link = element.querySelector("a[data-set-read]");
        let sendRequest = !link.dataset.noRequest;
        if(sendRequest){
            let entryID = parseInt(element.dataset.id, 10);
            this.updateEntriesStatus([entryID], "read");
        }

        link = element.querySelector("a[data-toggle-status]");
        if (link && link.dataset.value === "unread") {
            link.innerHTML = link.dataset.labelUnread;
            link.dataset.value = "read";
        }

        if (element && element.classList.contains("item-status-unread")) {
            element.classList.remove("item-status-unread");
            element.classList.add("item-status-read");
            UnreadCounterHandler.decrement(1);
        }
    }

    static setEntriesAboveStatusRead(element){
        let currentItem = document.querySelector(".current-item");
        let items = DomHelper.getVisibleElements(".items .item");
        if (currentItem === null || items.length === 0) {
            return;
        }
        let targetItems=[];
        let entryIds=[];
        for (let i = 0; i < items.length; i++) {
            targetItems.push(items[i]);
            entryIds.push(parseInt(items[i].dataset.id, 10));
            if (items[i].classList.contains("current-item")) {
                break;
            }
        }
        this.updateEntriesStatus(entryIds, "read",() => {
            targetItems.map(item => {
                let link = item.querySelector("a[data-toggle-status]");
                if (link && link.dataset.value === "unread") {
                    link.innerHTML = link.dataset.labelUnread;
                    link.dataset.value = "read";
                }
                if (item && item.classList.contains("item-status-unread")) {
                    item.classList.remove("item-status-unread");
                    item.classList.add("item-status-read");
                    UnreadCounterHandler.decrement(1);
                }
            });
        });
    }

    static toggleBookmark(element) {
        element.innerHTML = element.dataset.labelLoading;

        let item = DomHelper.findParent(element, 'item');
        let request = new RequestBuilder(element.dataset.bookmarkUrl);
        request.withCallback(() => {
            if (element.dataset.value === "star") {
                element.innerHTML = element.dataset.labelStar;
                element.dataset.value = "unstar";
                if (item) item.classList.remove("item-starred");
            } else {
                element.innerHTML = element.dataset.labelUnstar;
                element.dataset.value = "star";
                if (item) item.classList.add("item-starred");
            }
        });
        request.execute();
    }

    static toggleCache(element) {
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

    static markEntryAsRead(element) {
        if (element.classList.contains("item-status-unread")) {
            element.classList.remove("item-status-unread");
            element.classList.add("item-status-read");

            let entryID = parseInt(element.dataset.id, 10);
            this.updateEntriesStatus([entryID], "read");
            UnreadCounterHandler.decrement(1);
        }
    }

    static saveEntry(element) {
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

    static fetchOriginalContent(element) {
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
}
