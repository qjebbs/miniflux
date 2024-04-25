class ActionMenu {
    static initialize(element) {
        let currentEntry = findEntry(element);
        if (!currentEntry) return;
        if (isListView()) {
            this.deselectEntries();
            this.highlightEntry(currentEntry);
        }

        ModalHandler.close();
        let template = document.getElementById("action-menus");
        if (template === null) return;

        let container;
        if (isListView()) {
            // menu for entries
            container = initMenu(currentEntry.querySelectorAll(".item-meta :is(a, button)"), "entries");
            document.querySelector("#menu-mark-above-read").addEventListener("click", () => {
                setEntriesAboveStatusRead(currentEntry);
                ModalHandler.close();
                this.deselectEntries();
            });
        } else if (currentEntry.classList.contains("entry")) {
            // menu for entry
            container = initMenu(currentEntry.querySelectorAll(".entry-actions :is(a, button)"), "entry");
        }
        // cancel menu
        document.querySelector("#menu-action-cancel").addEventListener("click", () => {
            ModalHandler.close();
            this.deselectEntries();
        });

        let modal = container.querySelector(".modal");
        if (modal === null) return;

        // initMenu creates menu for given links in action modal.
        // dataForValue specifies the part of predefined menu to keep, 
        // which have the given value for "data-for" attribute.
        function initMenu(links, dataForValue) {
            let container = ModalHandler.open(template.content);
            let list = document.querySelector(".action-menus #element-links");
            while (list.hasChildNodes()) {
                list.removeChild(list.firstChild);
            }

            links.forEach(
                link => {
                    if (link.dataset.actionMenuExcluded !== undefined) return;
                    let menu = document.createElement("li");
                    menu.innerText = link.innerText;
                    menu.addEventListener("click", () => {
                        clickElement(link);
                        ModalHandler.close();
                    });
                    list.appendChild(menu);
                }
            );

            document.querySelectorAll(".action-menus li[data-for]")
                .forEach(e => {
                    if (e.dataset.for !== dataForValue)
                        e.style.display = 'none';
                });
            return container;
        }

        function clickElement(element) {
            element.dispatchEvent(new Event(
                "click",
                { bubbles: true, cancelable: true, composed: true },
            ));
        }
    }
    static close() {
        this.deselectEntries();
        ModalHandler.close();
    }
    static switch(element) {
        let lastEntry = document.querySelector(".item.current-item");
        let currentEntry = findEntry(element);
        let menu = document.querySelector("#modal-container .modal");
        if (menu && lastEntry === currentEntry) {
            this.close();
        } else {
            this.initialize(element);
        }
    }
    static highlightEntry(entry) {
        entry.classList.add("current-item");
    }
    static deselectEntries() {
        document.querySelectorAll(".current-item")
            .forEach(e => e.classList.remove("current-item"));
    }
}