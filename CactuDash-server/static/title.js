function title() {

    document.title = "CactuDash";

    var link = document.createElement("link");
    link.rel = "icon";
    link.href = "static/images/logomark.svg";

    
    document.head.appendChild(link);
}

title();
