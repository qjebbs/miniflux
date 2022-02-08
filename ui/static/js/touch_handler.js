class TouchHandler {
    constructor(className) {
        this.className = className;
        this.listeners = {
            start: [],
            move: [],
            active: [],
            end: []
        }
        this.reset();
        this.addEventListener("move", (e) => {
            if (!this.touch.element) return;
            this.touch.element.style.opacity = 1 - e.phase;
            this.touch.element.style.transform = "translateX(" + e.offset + "px)";
        });
    }

    reset() {
        this.touch = {
            start: { x: -1, y: -1 },
            move: { x: -1, y: -1 },
            element: null,
            flagDrag: false
        };
    }

    calculateDistance() {
        return this.touch.move.x - this.touch.start.x;
    }

    detectDrag() {
        if (this.touch.flagDrag) return;
        if (this.touch.start.x >= -1 && this.touch.move.x >= -1) {
            let horizontalDistance = Math.abs(this.touch.move.x - this.touch.start.x);
            let verticalDistance = Math.abs(this.touch.move.y - this.touch.start.y);

            if (horizontalDistance > 30 && verticalDistance < 70) {
                this.touch.flagDrag = true;
                this.runListeners(this.listeners.start);
            }
        }
    }

    findElement(element) {
        if (element.classList.contains(this.className)) {
            return element;
        }

        return DomHelper.findParent(element, this.className);
    }

    onTouchStart(event) {
        if (event.touches === undefined || event.touches.length !== 1) {
            return;
        }

        this.reset();
        this.touch.start.x = event.touches[0].clientX;
        this.touch.start.y = event.touches[0].clientY;
        this.touch.element = this.findElement(event.touches[0].target);
    }

    onTouchMove(event) {
        if (event.touches === undefined || event.touches.length !== 1 || this.element === null) {
            return;
        }

        this.touch.move.x = event.touches[0].clientX;
        this.touch.move.y = event.touches[0].clientY;
        this.detectDrag();

        if (this.touch.flagDrag) {
            this.runListeners(this.listeners.move);
            event.preventDefault();
        }
    }

    onTouchEnd(event) {
        if (event.touches === undefined) {
            return;
        }

        if (this.touch.element !== null) {
            let distance = this.calculateDistance();

            if (Math.abs(distance) > 75) {
                this.runListeners(this.listeners.active);
            } else {
                this.runListeners(this.listeners.end);
            }

            // If not on the unread page, undo transform of the dragged element.
            if (document.URL.split("/").indexOf("unread") == -1 || distance <= 75) {
                this.touch.element.style.opacity = 1;
                this.touch.element.style.transform = "none";
            }
        }
        this.reset();
    }

    listen() {
        let elements = document.querySelectorAll('.' + this.className);
        let hasPassiveOption = DomHelper.hasPassiveEventListenerOption();

        elements.forEach((element) => {
            element.addEventListener("touchstart", (e) => this.onTouchStart(e), hasPassiveOption ? { passive: true } : false);
            element.addEventListener("touchmove", (e) => this.onTouchMove(e), hasPassiveOption ? { passive: false } : false);
            element.addEventListener("touchend", (e) => this.onTouchEnd(e), hasPassiveOption ? { passive: true } : false);
            element.addEventListener("touchcancel", () => this.reset(), hasPassiveOption ? { passive: true } : false);
        });
    }
    addEventListener(type, listener) {
        switch (type) {
            case "active":
                this.listeners.active.push(listener);
                break;
            case "move":
                this.listeners.move.push(listener);
                break;
            case "start":
                this.listeners.start.push(listener);
                break;
            case "end":
                this.listeners.end.push(listener);
                break;
            default:
                break;
        }
    }
    runListeners(listeners) {
        listeners.forEach(fn => {
            let distance = this.calculateDistance();
            if (distance > 75) {
                distance = 75;
            } else if (distance < -75) {
                distance = -75;
            }
            let direction = distance > 0 ? "right" : "left";
            fn({
                target: this.touch.element,
                direction: direction,
                offset: distance,
                phase: Math.abs(distance) / 75,
            });
        });
    }
}

function initTouchHandlers() {
    let touchHandler = new TouchHandler('touch-item');
    touchHandler.addEventListener("start", (e) => {
        ActionMenu.close();
    });
    touchHandler.addEventListener("move", (e) => {
        if (e.direction == "left") {
            let menu = document.querySelector("#modal-container .modal");
            if (menu) {
                menu.style.transform = "translateX(" + (1 - e.phase) * 100 + "%)";
            } else {
                ActionMenu.initialize(e.target);
            }
        } else {
            ActionMenu.close();
        }
    });
    touchHandler.addEventListener("active", (e) => {
        if (e.direction == "right") {
            toggleEntryStatus(e.target);
        }
    });
    touchHandler.addEventListener("end", (e) => ActionMenu.close());
    touchHandler.listen();

    let entryContentElement = document.querySelector(".entry-content");
    if (entryContentElement) {
        let hasPassiveOption = DomHelper.hasPassiveEventListenerOption();
        let doubleTapTimers = {
            previous: null,
            next: null
        };

        const detectDoubleTap = (doubleTapTimer, event) => {
            const timer = doubleTapTimers[doubleTapTimer];
            if (timer === null) {
                doubleTapTimers[doubleTapTimer] = setTimeout(() => {
                    doubleTapTimers[doubleTapTimer] = null;
                }, 200);
            } else {
                event.preventDefault();
                goToPage(doubleTapTimer);
            }
        };

        entryContentElement.addEventListener("touchend", (e) => {
            if (e.changedTouches[0].clientX >= (entryContentElement.offsetWidth / 2)) {
                detectDoubleTap("next", e);
            } else {
                detectDoubleTap("previous", e);
            }
        }, hasPassiveOption ? { passive: false } : false);

        entryContentElement.addEventListener("touchmove", (e) => {
            Object.keys(doubleTapTimers).forEach(timer => doubleTapTimers[timer] = null);
        });
    }
}