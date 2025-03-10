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
            if (!this.touch.target) return;
            this.touch.target.style.opacity = 1 - e.phase;
            this.touch.target.style.transform = "translateX(" + e.offset + "px)";
        });
    }

    reset() {
        this.touch = {
            start: { x: -1, y: -1 },
            move: { x: -1, y: -1 },
            target: null,
            flagDrag: false,
            time: 0
        };
    }

    calculateDistance() {
        return this.touch.move.x - this.touch.start.x;
    }

    detectDrag() {
        if (this.touch.flagDrag) return;
        if (this.touch.start.x >= -1 && this.touch.move.x >= -1) {
            const horizontalDistance = Math.abs(this.touch.move.x - this.touch.start.x);
            const verticalDistance = Math.abs(this.touch.move.y - this.touch.start.y);

            if (horizontalDistance > 30 && verticalDistance < 70) {
                this.touch.flagDrag = true;
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
        this.touch.target = this.findElement(event.touches[0].target);
        this.touch.time = Date.now();
        this.runListeners(this.listeners.start);
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

        if (this.touch.target !== null) {
            const distance = this.calculateDistance();

            if (Math.abs(distance) > 75) {
                this.runListeners(this.listeners.active);
            } else {
                this.runListeners(this.listeners.end);
            }

            // If not on the unread page, undo transform of the dragged element.
            if (document.URL.split("/").indexOf("unread") == -1 || distance <= 75) {
                this.touch.target.style.opacity = 1;
                this.touch.target.style.transform = "none";
            }
        }
        this.reset();
    }

    listen() {
        let elements = document.querySelectorAll('.' + this.className);
        const eventListenerOptions = { passive: true };

        elements.forEach((element) => {
            element.addEventListener("touchstart", (e) => this.onTouchStart(e),eventListenerOptions);
            element.addEventListener("touchmove", (e) => this.onTouchMove(e));
            element.addEventListener("touchend", (e) => this.onTouchEnd(e),eventListenerOptions);
            element.addEventListener("touchcancel", () => this.reset(),eventListenerOptions);
        });
    }
    addEventListener(type, listener) {
        switch (type) {
            case "start":
                this.listeners.start.push(listener);
                break;
            case "move":
                this.listeners.move.push(listener);
                break;
            case "end":
                this.listeners.end.push(listener);
                break;
            case "active":
                this.listeners.active.push(listener);
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
                touch: this.touch,
                direction: direction,
                offset: distance,
                phase: Math.abs(distance) / 75,
            });
        });
    }
}

function initTouchHandlers() {
    let touchHandler = new TouchHandler('entry-swipe');
    touchHandler.addEventListener("start", (e) => {
        ActionMenu.close();
    });
    touchHandler.addEventListener("move", (e) => {
        if (e.direction == "left") {
            let menu = document.querySelector("#modal-container .modal");
            if (menu) {
                menu.style.transform = "translateX(" + (1 - e.phase) * 100 + "%)";
            } else {
                ActionMenu.initialize(e.touch.target);
            }
        } else {
            ActionMenu.close();
        }
    });
    touchHandler.addEventListener("active", (e) => {
        if (e.direction == "right") {
            toggleEntryStatus(e.touch.target);
        }
    });
    touchHandler.addEventListener("end", (e) => ActionMenu.close());
    touchHandler.listen();

    let entryContentElement = document.querySelector(".entry-content");
    if (entryContentElement) {
        let touchHandler = new TouchHandler('entry-content');
        if (!entryContentElement.classList.contains("gesture-nav-swipe")) {
            // action menu available
            touchHandler.addEventListener("move", (e) => {
                if (e.direction == "left") {
                    let menu = document.querySelector("#modal-container .modal");
                    if (menu) {
                        menu.style.transform = "translateX(" + (1 - e.phase) * 100 + "%)";
                    } else {
                        ActionMenu.initialize(e.touch.target);
                    }
                } else {
                    ActionMenu.close();
                }
            });
            touchHandler.addEventListener("end", (e) => ActionMenu.close());
        }
        if (entryContentElement.classList.contains("gesture-nav-tap")) {
            let lastTime = 0;
            touchHandler.addEventListener("start", (e) => {
                let now = Date.now();
                if (now - lastTime > 200) {
                    lastTime = now;
                    return;
                }
                if (e.touch.start.x >= entryContentElement.offsetWidth / 2){
                    goToPage("next");
                } else {
                    goToPage("previous");
                }
            });
        } else if (entryContentElement.classList.contains("gesture-nav-swipe")) {
            touchHandler.addEventListener("active", (e) => {
                if (e.direction == "left") {
                    goToPage("next");
                } else {
                    goToPage("previous");
                }
            });
        }
        touchHandler.listen();
    }
}