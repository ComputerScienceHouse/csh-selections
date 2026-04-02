function formSubmit(event) {
    event.preventDefault();
    let form = event.target;
    if (!form.reportValidity()) {
        return
    }
    let request = new XMLHttpRequest();
    request.open(form.method, form.action, true);
    request.onload = function () { // request successful
        // we can use server response to our request now
        console.log(request.response)
        if (200 <= request.status && request.status <= 210) {
            location.reload()
        } else {
            let res = request.responseText;
            try {
                res = JSON.parse(res).error;
            } catch (e) {
            }
            displayError(res)
            console.log(request.responseText)
        }
    };

    request.send(new FormData(form)); // create FormData from form that triggered event
}

function displayError(message) {
    let alertHolder = document.getElementById("alert-holder");
    alertHolder.innerHTML += `
    <div class="alert alert-warning alert-dismissible fade" role="alert">
        <span>${message}</span>
        <button type="button" class="btn-close" data-bs-dismiss="alert" aria-label="Close"></button>
    </div>
        `
    setTimeout(function () {
        alertHolder.lastElementChild.classList.add("show")
    }, .5)
}