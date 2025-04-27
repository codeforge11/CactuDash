// function toHHMMSS(secNum) {
//     var hours   = Math.floor(secNum / 3600);
//     var minutes = Math.floor((secNum - (hours * 3600)) / 60);
//     var seconds = secNum - (hours * 3600) - (minutes * 60);

//     if (hours   < 10) {hours   = "0"+hours;}
//     if (minutes < 10) {minutes = "0"+minutes;}
//     if (seconds < 10) {seconds = "0"+seconds;}
//     return hours+':'+minutes+':'+seconds;
// }

const socket = new WebSocket('ws://localhost:3030/ws');

socket.onmessage = function (event) {
    const data = JSON.parse(event.data);
    document.getElementById('processor').textContent = data.cpu_usage.toFixed(2) + '%';
};

socket.onopen = function () {
    console.log('WebSocket connection established');
    Log(false, "WebSocket connection established");
};

socket.onclose = function () {
    console.log('WebSocket connection closed');
    Log(false,"WebSocket connection closed");
};

socket.onerror = function (error) {
    console.log('WebSocket error:', error);
};

async function SysInfo() {
    try {
        const response = await fetch('/system-info');
        const data = await response.json();
        document.getElementById('hostname').innerText = data.hostname;
        document.getElementById('arch').innerText = data.arch;

        var supportStatus = data.supportStatus;       

        if (supportStatus){
            document.getElementById('OS_name').innerText = data.nameOfOs;
        }
        else{
            console.log("Error: Unsupported system!","\n","Please visit: https://github.com/codeforge11/CactuDash/wiki/Prerequisites")
            document.getElementById('OS_name').innerText = data.nameOfOs;
        }

        var lastTag = await fetch('/lastTag'); // Take last Tag from github
        var cactuCurrentVersion = await fetch('/cactu-dash'); // Current version
        const versionElement = document.getElementById('CactuDash_version'); //Place version on CactuDash_version

        const lastTagData = await lastTag.json();
        const cactuCurrentVersionData = await cactuCurrentVersion.json();

        if (lastTagData.version && lastTagData.version > cactuCurrentVersionData.version) {
            versionElement.style.color = 'red';
            versionElement.title = 'A new version is available';
        }
        
        versionElement.innerText = cactuCurrentVersionData.version;

        // Fetch disk usage
        const diskUsageResponse = await fetch('/disk-usage');
        const diskUsageData = await diskUsageResponse.json();

        const usedDiskMB = (diskUsageData.used / (1024 * 1024)).toFixed(2); // Convert to MB
        const totalDiskMB = (diskUsageData.total / (1024 * 1024)).toFixed(2); // Convert to MB
        const freeDiskMB = (diskUsageData.free / (1024 * 1024)).toFixed(2); // Convert to MB
        
        const usedDiskGB = (diskUsageData.used / (1024 * 1024 * 1024)).toFixed(2); // Convert to GB
        const totalDiskGB = (diskUsageData.total / (1024 * 1024 * 1024)).toFixed(2); // Convert to GB
        const freeDiskGB = (diskUsageData.free / (1024 * 1024 * 1024)).toFixed(2); // Convert to GB

        // Calculate percentage usage
        const usagePercentage = ((diskUsageData.used / diskUsageData.total) * 100).toFixed(2);

        // Update the text and progress bar
        updateDiskUsage(usedDiskMB, totalDiskMB, usedDiskGB, totalDiskGB, freeDiskMB, freeDiskGB, usagePercentage);
    } catch (error) {
        console.error('Error:', error);
    }
}

function updateDiskUsage(usedMB, totalMB, usedGB, totalGB, freeMB, freeGB, percentage) {
    // Update the progress bar
    const progressBar = document.getElementById('disk-progress');
    progressBar.style.width = percentage + '%';

    // Update the disk usage text
    const diskUsageText = document.getElementById('disk-usage');
    diskUsageText.innerText = `Total: ${totalGB} GB (${totalMB} MB) [${percentage}%] / Used: ${usedGB} GB (${usedMB} MB) / Free: ${freeGB} GB (${freeMB} MB) `;

   
    if (parseFloat(freeGB) < 10) {
        diskUsageText.style.color = 'red';
        diskUsageText.innerText += 'LOW FREE MEMORY !!! ';
        console.log("Low free memory lewel")
    } else {
        diskUsageText.style.color = 'white';
    }
}

async function reboot() {
    try {
        const response = await fetch('/power', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ option: false })
        });
        const data = await response.json();
        console.log(data);
    } catch (error) {
        console.error('Error:', error);
    }
}

async function shutdown() {
    try {
        const response = await fetch('/power', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ option: true })
        });
        const data = await response.json();
        console.log(data);
    } catch (error) {
        console.error('Error:', error);
    }
}

// Update function
async function update() {
    try {
        const response = await fetch('/update', { method: "POST" });
        const data = await response.json();
        console.log(data);
    } catch (error) {
        console.error('Error:', error);
    }
}

function openDocs() {
    window.open("https://github.com/codeforge11/CactuDash/wiki", '_blank').focus(); 
}

function CpuUsage(percentage) {
    const canvas = document.getElementById('cpuCanvas');
    const context = canvas.getContext('2d');
    const centerX = canvas.width / 2;
    const centerY = canvas.height / 2;
    const radius = Math.min(centerX, centerY) - 5;
    const endAngle = (percentage / 100) * 2 * Math.PI;
    context.clearRect(0, 0, canvas.width, canvas.height);
    context.beginPath();
    context.arc(centerX, centerY, radius, 0, 2 * Math.PI, false);
    context.fillStyle = '#e0e0e0';  
    context.fill();
    context.closePath();

    context.beginPath();
    context.moveTo(centerX, centerY);
    context.arc(centerX, centerY, radius, -0.5 * Math.PI, endAngle - 0.5 * Math.PI, false);
    context.fillStyle = '#007bff';  
    context.fill();
    context.closePath();
    context.beginPath();
    context.arc(centerX, centerY, radius - 12, 0, 2 * Math.PI, false);  
    context.fillStyle = '#ffffff'; 
    context.fill();
    context.closePath();

    context.shadowColor = "rgba(0, 0, 0, 0.3)";
    context.shadowBlur = 10;
    context.shadowOffsetX = 2;
    context.shadowOffsetY = 2;
}

socket.onmessage = function (event) {
    const data = JSON.parse(event.data);
    const cpuUsage = data.cpu_usage.toFixed(2);

    document.getElementById('cpuPercentage').textContent = cpuUsage + '%';

    CpuUsage(cpuUsage);
};

document.addEventListener('DOMContentLoaded', function () {
    fetch('/containers')
        .then(response => {
            if (!response.ok) {
                throw new Error('Network response was not ok');
            }
            return response.json();
        })
        .then(data => {
            if (data) {
                const table = document.getElementById('dockerTable');
                data.forEach(container => {
                    let row = table.insertRow();
                    
                    row.insertCell(0).innerText = container.Id;
                    row.insertCell(1).innerText = container.Name;
                    row.insertCell(2).innerText = container.Image;
                    row.insertCell(3).innerText = container.Status;

                    let actionsCell = row.insertCell(4);
                    let toggleButton = document.createElement('button');
                    
                    toggleButton.innerText = container.Status.includes("Up") ? 'Stop' : 'Start';
                    toggleButton.onclick = function () {
                        fetch('/toggle/' + container.Id, { method: 'POST' })
                            .then(() => location.reload());
                    };

                    actionsCell.appendChild(toggleButton);
                });
            }
        })
        .catch(error => {
            console.error('Error fetching containers:', error);
        });
});

function Log(type, message) {

    if (typeof type !== 'boolean') {
        console.error('Invalid log type:', type);
        return;
    }

    //for error log type must be true
    
    fetch('/log', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({ type: type, message: message })
    }).catch(error => {
        console.error('Error sending log:', error);
    });
};

SysInfo();
