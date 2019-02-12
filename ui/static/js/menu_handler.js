class MenuHandler {
    clickMenuListItem(event) {
        let element = event.target;

        if (element.tagName === "A") {
            window.location.href = element.getAttribute("href");
        } else {
            window.location.href = element.querySelector("a").getAttribute("href");
        }
    }

    logoClickHandler(event) {
        if (document.documentElement.clientWidth < 600)
            this.toggleMainMenu();
        else
            this.clickMenuListItem(event);
    }


    toggleMainMenu() {
        let menu = document.querySelector(".header nav ul");
        if (DomHelper.isVisible(menu)) {
            menu.classList.remove("show");
        } else {
            menu.classList.add("show");
        }

        let searchElement = document.querySelector(".header .search");
        if (DomHelper.isVisible(searchElement)) {
            searchElement.classList.remove("show");
        } else {
            searchElement.classList.add("show");
        }
    }
}
