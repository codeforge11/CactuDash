// function toHHMMSS(secNum) {
//     var hours   = Math.floor(secNum / 3600);
//     var minutes = Math.floor((secNum - (hours * 3600)) / 60);
//     var seconds = secNum - (hours * 3600) - (minutes * 60);

//     if (hours   < 10) {hours   = "0"+hours;}
//     if (minutes < 10) {minutes = "0"+minutes;}
//     if (seconds < 10) {seconds = "0"+seconds;}
//     return hours+':'+minutes+':'+seconds;
// }



async function SysInfo() {
    try {
        const response = await fetch('/os-name');
        const data = await response.json();
        document.getElementById('os-name').innerText = data.os_name;

        // const uptimeResponse = await fetch('/uptime');
        // const uptimeData = await uptimeResponse.json();
        
        // document.getElementById('uptime').innerText = toHHMMSS(uptimeData.uptime);

    } catch (error) {
        console.error('Error:', error);
    }
}

SysInfo(); 
