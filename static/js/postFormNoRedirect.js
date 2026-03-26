function formSubmit(event) {
    event.preventDefault();
    let form = event.target;
    let request = new XMLHttpRequest();
    request.open(form.method, form.action, true);
    request.onload = function () { // request successful
        // we can use server response to our request now
        console.log(request.response)
        if (200 <= request.status && request.status <= 210) {
            location.reload()
        } else {
            console.log(request.responseText)
            //TODO: alert this in a cool js modal
        }
    };

    request.send(new FormData(form)); // create FormData from form that triggered event
}