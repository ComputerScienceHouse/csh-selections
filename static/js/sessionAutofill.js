import "@selectize/selectize";

async function onLoad() {
    let input = document.getElementById("membersSelect")

    await fetch("/session/allMembers", {
        method: "GET",
        headers: {
            'Accept': "application/json",
            'Content-Type': "application/json"
        }
    }).then(res => res.json())
        .then((data) => this.members = data)

    input.selectize({
        plugins: ['remove_button'],
        persist: false,
        openOnFocus: false,
        closeAfterSelect: true,
        valueField: 'value',
        labelField: 'display',
        searchField: 'display',
        selectOnTab: true,
        options: this.members
    })
}

document.addEventListener('load', onLoad)