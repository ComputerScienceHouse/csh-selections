async function onLoad() {
    // load CSS
    let selectizeRef = document.createElement('link');
    selectizeRef.rel = "stylesheet"
    selectizeRef.href = "https://raw.githubusercontent.com/CarstenHager/selectize.js/2df8e8ea6aa9cdf3f3c33126645adfc6e6efb327/dist/css/selectize.bootstrap5-dark-mode.css"
    document.head.appendChild(selectizeRef)

    let input = document.getElementById("membersSelect")

    await fetch("/session/allMembers", {
        method: "GET",
        headers: {
            'Accept': "application/json",
            'Content-Type': "application/json"
        }
    }).then(res => res.json())
        .then((data) => this.members = data.members)

    $(input).selectize({
        plugins: ['remove_button'],
        persist: false,
        openOnFocus: false,
        closeAfterSelect: true,
        valueField: 'Username',
        labelField: 'Name',
        searchField: ['Username', 'Name'],
        selectOnTab: true,
        options: this.members
    })
}

document.addEventListener('load', onLoad)
onLoad()