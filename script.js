var tree = null;

function addToTree(container, node) {
    function walk(n) {
        if (n.Element == container) {
            n.Children.push(node);
            return;
        }
        if (n.Children) {
            for (let i = 0; i < n.Children.length; i++) {
                walk(n.Children[i]);
            }
        }
    }

    walk(tree);
}

var prefixNameToSize = {
    "uint8 prefixed": 1,
    "uint16 prefixed": 2,
}

var inputID = 0;

function addValue(event) {
    button = event.currentTarget;
    dropdown = document.getElementById(button.id+"-select");

    
    if (dropdown.value == "uint8 prefixed" || dropdown.value == "uint16 prefixed") {
        parent = button.parentElement;
        addContainer(parent, dropdown, dropdown.value, prefixNameToSize[dropdown.value]);
        return
    }
    
    val = document.createElement("div");
    val.classList.add("value");
    label = document.createElement("div");
    label.classList.add("label");
    label.innerText = dropdown.value;
    val.appendChild(label);
    input = document.createElement("input");
    input.id = "input-"+inputID;
    input.classList.add("input");
    val.appendChild(input);

    if (button.parentElement != null) {
        addToTree(button.parentElement, {Element: input, Type: dropdown.value});
    }

    dropdown.parentElement.insertBefore(val, dropdown);

    inputID++;
}

var selectorID = 0;

function addAdder(element) {
    const values = ["uint8", "uint16", "string", "hex bytes", "uint8 prefixed", "uint16 prefixed"];
    select = document.createElement("select");
    select.id = "adder-"+selectorID+"-select";
    values.forEach((value) => {
        option = document.createElement("option");
        option.value = value;
        option.innerText = value;
        select.appendChild(option);
    })
    element.appendChild(select);
    button = document.createElement("button");
    button.id = "adder-"+selectorID;
    button.innerText = "add";
    button.addEventListener('click', addValue, false);
    element.appendChild(button);

    selectorID++;
}

var containerID = 0;

function addContainer(element, before, prefixName, prefixSize) {
    container = document.createElement("div");
    container.id = "container-"+containerID;
    container.classList.add("container");
    if (prefixName != "") {
        label = document.createElement("div");
        label.innerText = prefixName;
        container.appendChild(label);
    }

    if (tree != null) {
        addToTree(element, {Element: container, Children: [], PrefixSize: prefixSize});
    } else {
        tree = {Element: container, Children: []};
    }

    addAdder(container);
    element.insertBefore(container, before);

    containerID++;
}

function init() {
    addContainer(document.body, null, "", 0);
    builder = document.createElement("button");
    builder.innerText = "build";
    builder.addEventListener('click', callBuild, false);
    document.body.appendChild(builder);
    hex = document.createElement("div");
    hex.id = "hex";
    hex.classList.add("hex");
    document.body.appendChild(hex);
}

function callBuild(event) {
    j = [];

    function walk(n) {
        if (n.PrefixSize > 0) {
            o = {
                Tag: n.Element.id,
                Type: "prefix",
                PrefixSize: n.PrefixSize,
                Children: [],
            };
            for (let x = 0; x < n.Children.length; x++) {
                o.Children.push(walk(n.Children[x]));
            }
            return o;
        } else {
            return {
                Tag: n.Element.id,
                Type: "concrete",
                ValueType: n.Type,
                Value: n.Element.value,
            };
        }
    }

    for (let i = 0; i < tree.Children.length; i++) {
        j.push(walk(tree.Children[i]));
    }

    result = build(JSON.stringify(j));

    document.getElementById("hex").innerText = [...result.hex].map((d, i) => (i) % 2 == 0 ? ' ' + d : d).join('').trim();
}