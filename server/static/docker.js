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
    <div id="dockerCreatePanel">
        <div style="display: flex; align-items: center; margin-bottom: 16px;">
            <span style="font-size: 2rem; font-weight: bold; color: #fff;">Docker Run</span>
        </div>
        <textarea
            class="w-full max-w-4xl bg-[#23283a] rounded-lg shadow-lg px-6 py-4 text-[#e0e0e0] font-mono text-lg outline-none"
            id="DockerCode" rows="10" placeholder="docker run ..." maxlength="1000"></textarea>
        <div style="display: flex; gap: 20px;">
            <button onclick="createDockerPush()" style="flex: 1; padding: 16px 0; font-size: 1.1rem; border-radius: 8px; background: #2496ed; color: #fff; border: none; font-weight: bold; cursor: pointer;">Create Docker Image</button>
            <button onclick="addContainerShowElements()" style="flex: 1; padding: 16px 0; font-size: 1.1rem; border-radius: 8px; background: #444950; color: #fff; border: none; font-weight: bold; cursor: pointer;">Cancel</button>
        </div>
    </div>
    `;

    const textarea = document.getElementById('DockerCode');
    textarea.addEventListener('input', function() {
        const maxRows = 10;
        const lines = textarea.value.split('\n');
        if (lines.length > maxRows) {
            textarea.value = lines.slice(0, maxRows).join('\n');
        }
    });
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
    try {
        const res = await fetch('/createDockerType', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ code, type: true})
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
