
const HTTP_API_URL = "http://127.0.0.1:9045"



// fetch the data from the API server
// returns a promise for the json data.
function fetchMsgHistory(reqBody){
    let responseJson = {}
    return fetch(HTTP_API_URL, {
        method: "POST",
        body: JSON.stringify(reqBody)
    })
    .then((res) => res.json())
}

export {
    fetchMsgHistory
}