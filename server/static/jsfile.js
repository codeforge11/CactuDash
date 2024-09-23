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
    
socket.onmessage = function(event) {
    const data = JSON.parse(event.data);
    document.getElementById('processor').textContent = data.cpu_usage.toFixed(2) + '%';
};

socket.onopen = function() {
    console.log('WebSocket connection established');
};

socket.onclose = function() {
    console.log('WebSocket connection closed');
};

socket.onerror = function(error) {
    console.log('WebSocket error:', error);
};

async function SysInfo() {
    try {
        const response = await fetch('/os-name');
        const data = await response.json();
        document.getElementById('os-name').innerText = data.os_name;

        const kernelResponse = await fetch('/kernel_version');
        const kernelData = await kernelResponse.json();
        document.getElementById('kernel').innerText = kernelData.os_name;

        // const uptimeResponse = await fetch('/uptime');
        // const uptimeData = await uptimeResponse.json();
        
        // document.getElementById('uptime').innerText = toHHMMSS(uptimeData.uptime);

    } catch (error) {
        console.error('Error:', error);
    }
}

SysInfo(); 
