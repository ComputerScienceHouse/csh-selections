function formSubmit(event) {
    event.preventDefault();
    let form = event.target;
    let request = new XMLHttpRequest();
    request.open('POST', form.action, true);
    request.onload = function () { // request successful
        // we can use server response to our request now
    };

    request.onerror = function () {
        console.log(request.responseText);
        // request failed
    };

    request.send(new FormData(form)); // create FormData from form that triggered event
}

// and you can attach form submit event like this for example
function attachFormSubmitEvent(formId) {
    document.getElementById(formId).addEventListener("submit", formSubmit);
}