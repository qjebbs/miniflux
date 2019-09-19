class RequestBuilder {
    constructor(url) {
        this.callback = null;
        this.url = url;
        this.options = {};
    }

    withBody(body) {
        this.options = {
            method: "POST",
            cache: "no-cache",
            credentials: "include",
            body: JSON.stringify(body),
            headers: new Headers({
                "Content-Type": "application/json",
                "X-Csrf-Token": this.getCsrfToken()
            })
        };
        return this;
    }

    withForm(data) {
        let form = new FormData();
        for (let key in data) {
            form.append(key, data[key]);
        }
        this.options = {
            method: "POST",
            cache: "no-cache",
            credentials: "include",
            body: form,
            headers: new Headers({
                "X-Csrf-Token": this.getCsrfToken()
            })
        };
        return this;
    }

    withCallback(callback) {
        this.callback = callback;
        return this;
    }

    getCsrfToken() {
        let element = document.querySelector("meta[name=X-CSRF-Token]");
        if (element !== null) {
            return element.getAttribute("value");
        }

        return "";
    }

    execute() {
        fetch(new Request(this.url, this.options)).then((response) => {
            if (this.callback) {
                this.callback(response);
            }
        });
    }
}
