function handleActionMenu(element) {
    let currentEntry = findEntry(element);
    if (!currentEntry) return;
    // if (isListView()) highlightEntry(currentEntry);

    ModalHandler.close();
    let template = document.getElementById("action-menus");
    if (template === null) return;

    if (currentEntry.classList.contains("item")) {
        // menu for entries
        initMenu(currentEntry.querySelectorAll(".item-meta a"), "entries");
        document.querySelector("#menu-mark-above-read").addEventListener("click", () => {
            setEntriesAboveStatusRead(currentEntry);
            ModalHandler.close();
        });
    } else if (currentEntry.classList.contains("entry")) {
        // menu for entry
        initMenu(currentEntry.querySelectorAll(".entry-actions a"), "entry");
    }
    // cancel menu
    document.querySelector("#menu-action-cancel").addEventListener("click", () => {
        ModalHandler.close();
    });

    function highlightEntry(entry) {
        document.querySelectorAll(".current-item")
            .forEach(e => e.classList.remove("current-item"));
        entry.classList.add("current-item");
    }
    // initMenu creates menu for given links in action modal.
    // dataForValue specifies the part of predefined menu to keep, 
    // which have the given value for "data-for" attribute.
    function initMenu(links, dataForValue) {
        ModalHandler.open(template.content, true);
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
    }

    function clickElement(element) {
        let e = document.createEvent("MouseEvents");
        e.initEvent("click", true, true);
        element.dispatchEvent(e);
    }
}