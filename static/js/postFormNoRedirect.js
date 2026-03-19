function formSubmit(event) {
    event.preventDefault();
    let form = event.target;
    let request = new XMLHttpRequest();
    request.open(form.method, form.action, true);
    request.onload = function () { // request successful
        // we can use server response to our request now
        location.reload()
    };

    request.onerror = function () {
        console.log(request.responseText);
        // request failed
    };

    request.send(new FormData(form)); // create FormData from form that triggered event
}