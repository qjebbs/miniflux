class TabHandler {
    constructor() {
        this.listeners = {};
        document.querySelectorAll(".tabs").forEach(tabs => {
            this.add(tabs);
        });
    }
    add(tabs) {
        let self = this;
        let headers = tabs.querySelectorAll('.tab-head li');
        let contents = tabs.querySelectorAll('.tab-body .tab-content');
        if (!headers.length || !contents.length) return;

        this.listeners[tabs] = [];
        let showTab = function () {
            return function () {
                for (var i = 0, len = headers.length; i < len; i++) {
                    if (headers[i] === this) {
                        headers[i].classList.add('active');
                        contents[i].classList.add('active');
                        for (let listener of self.listeners[tabs]) {
                            listener.call(headers[i], headers[i], contents[i], i);
                        }
                    } else {
                        headers[i].classList.remove('active');
                        contents[i].classList.remove('active');
                    }
                }
            }
        }();
        headers.forEach(e => {
            e.addEventListener("click", showTab);
        });
    }
    addEventListener(tabs, listener) {
        if (typeof tabs === 'string') {
            tabs = document.querySelector(tabs);
        }
        if (!tabs || !this.listeners[tabs]) return;
        this.listeners[tabs].push(listener);
    }
}