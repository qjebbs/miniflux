class EntryEditorHandler {
    static switchHandler(header, content, i) {
        let preview = document.querySelector('#preview-content');
        let editor = document.querySelector('#form-content');
        if (i == 0) {
            editor.value = preview.innerHTML;
        } else {
            preview.innerHTML = editor.value;
        }
    }
    static submitHandler() {
        let preview = document.querySelector('#preview-content');
        let editor = document.querySelector('#form-content');
        let previewParent = DomHelper.findParent(preview, "tab-content");
        if (previewParent.classList.contains('active')) {
            editor.value = preview.innerHTML;
        }
        document.querySelector("#entry-form").submit();
    }
}