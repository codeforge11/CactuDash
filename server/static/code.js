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
};

socket.onclose = function () {
    console.log('WebSocket connection closed');
};

socket.onerror = function (error) {
    console.log('WebSocket error:', error);
};

async function SysInfo() {
    try {
        const response = await fetch('/system-info');
        const data = await response.json();
        document.getElementById('hostname').innerText = data.hostname;
        document.getElementById('kernel').innerText = data.kernel_version;

        const cactuDashResponse = await fetch('/cactu-dash');
        const cactuDashData = await cactuDashResponse.json();
        document.getElementById('CactuDash_version').innerText = cactuDashData.version;

        // Fetch disk usage
        const diskUsageResponse = await fetch('/disk-usage');
        const diskUsageData = await diskUsageResponse.json();

        const usedDiskMB = (diskUsageData.used / (1024 * 1024)).toFixed(2); // Convert to MB
        const totalDiskMB = (diskUsageData.total / (1024 * 1024)).toFixed(2); // Convert to MB
        const usedDiskGB = (diskUsageData.used / (1024 * 1024 * 1024)).toFixed(2); // Convert to GB
        const totalDiskGB = (diskUsageData.total / (1024 * 1024 * 1024)).toFixed(2); // Convert to GB

        // Calculate percentage usage
        const usagePercentage = ((diskUsageData.used / diskUsageData.total) * 100).toFixed(2);

        // Update the text and progress bar
        updateDiskUsage(usedDiskMB, totalDiskMB, usedDiskGB, totalDiskGB, usagePercentage);
    } catch (error) {
        console.error('Error:', error);
    }
}

function updateDiskUsage(usedMB, totalMB, usedGB, totalGB, percentage) {
    // Update the progress bar
    const progressBar = document.getElementById('disk-progress');
    progressBar.style.width = percentage + '%';

    // Update the disk usage text
    const diskUsageText = document.getElementById('disk-usage');
    diskUsageText.innerText = `Used: ${usedGB} GB (${usedMB} MB) / Total: ${totalGB} GB (${totalMB} MB) [${percentage}%]`;
}

async function reboot() {
    try {
        const response = await fetch('/reboot', { method: "POST" });
        const data = await response.json();
        console.log(data);
    } catch (error) {
        console.error('Error:', error);
    }
}

SysInfo();
