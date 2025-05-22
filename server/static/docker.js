function addContainerShowElements() {

    const addContainerCenter = document.getElementById("addContainerCenter");

    addContainerCenter.innerHTML = `
        <div id="dockerCreate" title="Create docker" onclick="createDocker()">
            <img src="static/img/docker/docker-logo.svg" alt="docker">
        </div>

        <div id="dockerComposeCreate" title="Create docker compose" onclick="createDockerCompose()">
            <img src="static/img/docker/docker-compose-logo.png" alt="docker compose">
        </div>
    `;
}

function createDocker() {
    document.getElementById("addContainerCenter").innerHTML = `
        <textarea id="DockerCode" rows="10" style="width: 100%;"></textarea>
        <input type="text" name="DockerImageName" id="DockerImageName">
        <button onclick="createDockerPush()">Create Docker Image</button>
        <button onclick="addContainerShowElements()">Cancel</button>
    `;
}

function createDockerCompose() {
        document.getElementById("addContainerCenter").innerHTML = `
        <textarea id="DockerComposeCode" rows="10" style="width: 100%;"></textarea>
        <button onclick="createDockerComposePush()">Create Docker Image</button>
        <button onclick="addContainerShowElements()">Cancel</button>
    `;
}

async function createDockerPush() {
    const code = document.getElementById('DockerCode').value;

    const name = document.getElementById('DockerImageName').value;

    if (name ==""){
        name = "image";
    }

    try {
        const res = await fetch('/createDockerType', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ code, type: true, name})
        });
        console.log(await res.json());
        addContainerShowElements();
    } catch (error) {
        console.error('Error:', error);
    }

}

async function createDockerComposePush() {
    const code = document.getElementById('DockerComposeCode').value;

    try {
        const res = await fetch('/createDockerType', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ code, type: false })
        });
        console.log(await res.json());
        addContainerShowElements();
    } catch (error) {
        console.error('Error:', error);
    }

}
