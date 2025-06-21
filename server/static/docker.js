function showElements() {

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
            id="DockerCode" rows="3" placeholder="docker run ..." maxlength="1000"></textarea>
        <div style="display: flex; gap: 20px;">
            <button onclick="createDockerPush()" class="dockerCrButtons" style="background: #2496ed;">Create Docker Image</button>
            <button onclick="showElements()" class="dockerCrButtons" style="background: #444950;">Cancel</button>
        </div>
    </div>
    `;

    const textarea = document.getElementById('DockerCode');
    textarea.addEventListener('input', function() {
        const maxRows = 3;
        const lines = textarea.value.split('\n');
        if (lines.length > maxRows) {
            textarea.value = lines.slice(0, maxRows).join('\n');
        }
    });
}

function createDockerCompose() {
    document.getElementById("addContainerCenter").innerHTML = `
        <form id="dockerCreatePanel" style="max-width: 800px; margin: 0 auto;" onsubmit="event.preventDefault(); createDockerComposePush();">
            <div style="display: flex; align-items: flex-start; margin-bottom: 16px;">
                <span style="font-size: 2rem; font-weight: bold; color: #fff;">compose.yaml</span>
            </div>
            <div style="text-align: left;">
                <h1 style="color: white;">Stack Name </h1>
                <input type="text" name="stackName" id="stackName" oninput="this.value = this.value.toLowerCase();">
                <p style="color: grey; font-size: 12px">Lowercase only </p>
            </div>
            <div style="display: flex; align-items: stretch; max-height: 400px; overflow: hidden; border-radius: 8px;">
                <div id="codeLineNumbers"></div>
                <textarea class="w-full bg-[#23283a] rounded-lg shadow-lg px-6 py-4 text-[#e0e0e0] font-mono text-lg outline-none"
                    id="DockerCode" name="DockerCode" rows="16" maxlength="10000"
                    style="padding-top: 2px; border-radius: 0 8px 8px 0; border-left: 1px solid #333; min-width: 400px; min-height: 320px; display: block; line-height: 1.5; overflow: auto; overflow-y: auto;" required></textarea>
            </div>
            <div style="display: flex; gap: 20px; margin-top: 16px;">
                <button type="submit" class="dockerCrButtons" style="background: #2496ed;">Create Docker Container</button>
                <button type="button" onclick="showElements()" class="dockerCrButtons" style="background: #444950;">Cancel</button>
            </div>
        </form>
    `;

    const textarea = document.getElementById('DockerCode');
    const lineNumbers = document.getElementById('codeLineNumbers');

    function updateLineNumbers() {
        const lines = textarea.value.split('\n').length || 1;
        lineNumbers.innerHTML = '';
        for (let i = 1; i <= lines; i++) {
            const line = document.createElement('div');
            line.textContent = i;
            line.style.height = getComputedStyle(textarea).lineHeight;
            line.style.lineHeight = getComputedStyle(textarea).lineHeight;
            lineNumbers.appendChild(line);
        }
        lineNumbers.style.height = textarea.clientHeight + 'px';
    }

    textarea.addEventListener('input', updateLineNumbers);
    textarea.addEventListener('scroll', function() {
        lineNumbers.scrollTop = textarea.scrollTop;
    });

    const style = getComputedStyle(textarea);
    lineNumbers.style.fontFamily = style.fontFamily;
    lineNumbers.style.fontSize = style.fontSize;
    lineNumbers.style.lineHeight = style.lineHeight;

    updateLineNumbers();
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
        showElements();
    } catch (error) {
        console.error('Error:', error);
    }
}

async function createDockerComposePush() {
    const code = document.getElementById('DockerCode').value;
    let name = document.getElementById('stackName').value;
    if (!name) {
        name = new Date().toISOString().replace(/[:.]/g, '-');
    }

    try {
        const res = await fetch('/createDockerType', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ code, type: false, name })
        });
        console.log(await res.json());
        showElements();
    } catch (error) {
        showElements()
        console.error('Error:', error);
    }
}