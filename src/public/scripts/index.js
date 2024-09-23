const grid = document.getElementById("wormo-grid");
const foodCounter = document.getElementById("food-counter");
const foodNeeded = document.getElementById("food-needed");
const progressBar = document.getElementById("progress-inner");
const loading = document.getElementById("ui-loading");
const progressBox = document.getElementById("ui-progress");

const TOTAL_SPACES = GRID_ROWS * GRID_ROWS;

const LEVEL_MULTIPLIER = 2;

const wormFoodLocations = new Set();

const calculateFoodToExtend = length => length * LEVEL_MULTIPLIER;
const checkCellInBounds = ({x, y}) => (x >= 0 && x < GRID_COLS) && (y >= 0 && y < GRID_ROWS);
const checkCellHasFood = ({x, y}) => wormFoodLocations.has(`${x}-${y}`);

const addColourToCell = ({x, y}, colour) => {
    const index = x + y * GRID_COLS;
    console.debug(`Adding colour ${colour} to ${x},${y}`)

    const cell = grid.children.item(index);
    cell.style["background-color"] = colour;
};

const removeColourFromCell = ({x, y}) => {
    const index = x + y * GRID_COLS;
    console.debug(`Removing colour from ${x},${y}`)

    const cell = grid.children.item(index);
    cell.style.removeProperty("background-color");
};

const addFoodToCell = ({x, y}, colour) => {
    const index = x + y * GRID_COLS;
    console.debug(`Adding worm food to ${x},${y}`)

    const cell = grid.children.item(index);
    cell.innerHTML = "&#x25CF;";
    cell.classList.add("worm-food");
    cell.style.color = colour;

    wormFoodLocations.add(`${x}-${y}`);
};

const removeFoodFromCell = ({x, y}) => {
    const index = x + y * GRID_COLS;
    console.debug(`Removing worm food from ${x},${y}`)

    const cell = grid.children.item(index);
    cell.innerHTML = "";
    cell.classList.remove("worm-food");
    cell.style.removeProperty("colour");

    wormFoodLocations.delete(`${x}-${y}`);
};

const checkCellIsEmpty = ({x, y}) => {
    const playerWormPresent = playerWorm.positions.some(pos => pos.x === x && pos.y === y);
    const inBounds = checkCellInBounds({x, y});

    return !playerWormPresent && inBounds;
};

const findAdjacentFreeSpace = ({x, y}) => {
    for(let i = -1; i <= 1; i++){
        const newX = x + i;

        for(let j = -1; j <= 1; j++){
            const newY = y + j;
            const newPos = {x: newX, y: newY};

            if(checkCellIsEmpty(newPos)){
                return newPos;
            }
        }
    }

    return null;
};

const updateFoodCounter = (consumed, needed) => {
    foodCounter.innerHTML = consumed;
    foodNeeded.innerHTML = needed;

    const percentage = consumed / needed * 100;

    progressBar.style.width = `${percentage}%`;
};

const generateRandomColour = () => '#' + (Math.random().toString(16) + "000000").substring(2,8);

class PlayerWorm {
    /**
        @positions {x: number, y: number}[] head->tail
        @colour string
        @headColour string
    **/
    constructor(positions, colour, headColour) {
        this.positions = positions;
        this.colour = colour;
        this.headColour = headColour;
        this.foodConsumed = 0;

        for(let i = 1; i < positions.length; i++){
            addColourToCell(positions[i], colour);
        }

        addColourToCell(positions[0], headColour);
    }

    /**
        @direction "U" | "D" | "L" | "R"
    **/
    move(direction) {
        let headPos = this.positions[0];
        const newHeadPos = {...headPos};

        switch(direction){
            case 'U':
                newHeadPos.y--;
                break;
            case 'D':
                newHeadPos.y++;
                break;
            case 'L':
                newHeadPos.x--;
                break;
            case 'R':
                newHeadPos.x++;
                break;
        }

        if(!checkCellInBounds(newHeadPos)){
            return;
        }

        removeColourFromCell(headPos);

        for(let i = this.positions.length - 1; i > 0; i--){
            let pos = this.positions[i];
            removeColourFromCell(pos);

            const nextPos = this.positions[i - 1];

            pos.x = nextPos.x;
            pos.y = nextPos.y;
        }

        for(let i = this.positions.length - 1; i > 0; i--){
            addColourToCell(this.positions[i], this.colour);
        }

        headPos.x = newHeadPos.x;
        headPos.y = newHeadPos.y;

        addColourToCell(headPos, this.headColour);

        if(checkCellHasFood(headPos)){
            removeFoodFromCell(headPos);

            this.foodConsumed++;

            console.debug("Consumed: " + this.foodConsumed);
            console.debug("Needed: " + calculateFoodToExtend(this.positions.length));

            if(this.foodConsumed === calculateFoodToExtend(this.positions.length)){
                this.extend();
            }

            updateFoodCounter(this.foodConsumed, calculateFoodToExtend(this.positions.length));
        }
    }

    extend() {
        const tailPos = this.positions[this.positions.length - 1];
        const newPos = findAdjacentFreeSpace(tailPos);

        this.positions.push(newPos);
        addColourToCell(newPos, this.colour);
        this.foodConsumed = 0;

        ws.send(wsEvents.EXTEND + "\n" + newPos.x + ":" + newPos.y);
    }

    getPositionString() {
        let ret = this.positions[0].x + ':' + this.positions[0].y;

        for(let i = 1; i < this.positions.length; i++){
            ret += ',' + this.positions[i].x + ':' + this.positions[i].y;
        }

        return ret;
    }
}

class EnemyWorm {
    /**
        @positions {x: number, y: number}[] head->tail
        @colour string
        @headColour string
    **/
    constructor(positions, colour, headColour) {
        this.positions = positions;
        this.colour = colour;
        this.headColour = headColour;

        for(let i = 1; i < positions.length; i++){
            addColourToCell(positions[i], colour);
        }

        addColourToCell(positions[0], headColour);
    }

    /**
        @direction "U" | "D" | "L" | "R"
    **/
    move(direction) {
        let headPos = this.positions[0];
        const newHeadPos = {...headPos};

        switch(direction){
            case 'U':
                newHeadPos.y--;
                break;
            case 'D':
                newHeadPos.y++;
                break;
            case 'L':
                newHeadPos.x--;
                break;
            case 'R':
                newHeadPos.x++;
                break;
        }

        if(!checkCellInBounds(newHeadPos)){
            return;
        }

        removeColourFromCell(headPos);

        for(let i = this.positions.length - 1; i > 0; i--){
            let pos = this.positions[i];
            removeColourFromCell(pos);

            const nextPos = this.positions[i - 1];

            pos.x = nextPos.x;
            pos.y = nextPos.y;
        }

        for(let i = this.positions.length - 1; i > 0; i--){
            addColourToCell(this.positions[i], this.colour);
        }

        headPos.x = newHeadPos.x;
        headPos.y = newHeadPos.y;

        addColourToCell(headPos, this.headColour);

        if(checkCellHasFood(headPos)){
            removeFoodFromCell(headPos);
        }
    }

    extend(newPos) {
        this.positions.push(newPos);
    }

    getPositionString() {
        let ret = this.positions[0].x + ':' + this.positions[0].y;

        for(let i = 1; i < this.positions.length; i++){
            ret += ',' + this.positions[i].x + ':' + this.positions[i].y;
        }

        return ret;
    }
}

addEventListener("keydown", (event) => {
    if(event.repeat){
        return;
    }

    let dir = null;

    if(event.key === "ArrowUp"){
        dir = 'U';
    } else if(event.key === "ArrowDown"){
        dir = 'D';
    } else if(event.key === "ArrowLeft"){
        dir = 'L';
    } else if(event.key === "ArrowRight"){
        dir = 'R';
    } else{
        return;
    }

    playerWorm.move(dir);
    ws.send(wsEvents.MOVE + "\n" + dir);
});

const parsePositions = (str) => {
    const splitStr = str.split(',');
    const positions = Array(splitStr.length);

    for(let i = 0; i < splitStr.length; i++){
        const [x, y] = splitStr[i].split(':');
        positions[i] = {x: parseInt(x), y: parseInt(y)};
    }

    return positions;
};

const parseNewEvent = (str) => {
    const firstComma = str.indexOf(',');
    const id = str.slice(0, firstComma);
    const positions = parsePositions(str.slice(firstComma + 1));

    return [id, positions];
};

let playerWorm;
let enemyWorms = new Map();
let foodPositions = new Set();

let ws;

const wsEvents = {
    EXTEND: "EXTEND",
    NEW: "NEW",
    MOVE: "MOVE",
    SPAWNFOOD: "SPAWNFOOD",
};

const handleWsMsg = ({ data }) => {
    console.debug("ws msg: " + data);

    const firstNewLine = data.indexOf("\n");

    const event = data.slice(0, firstNewLine);
    const msg = data.slice(firstNewLine + 1, data.length);

    switch(event){
        case wsEvents.NEW: {
            const [id, positions] = parseNewEvent(msg);

            const enemyWorm = new EnemyWorm(
                positions,
                generateRandomColour(),
                generateRandomColour(),
            );
            enemyWorms.set(id, enemyWorm);

            break;
        }
        case wsEvents.MOVE: {
            const [id, dir] = msg.split(',');
            enemyWorms.get(id).move(dir);

            break;
        }
        case wsEvents.SPAWNFOOD: {
            const [xStr, yStr] = msg.split(':');
            addFoodToCell({x: parseInt(xStr), y: parseInt(yStr)}, generateRandomColour());

            break;
        }
        case wsEvents.EXTEND: {
            const [id, pos] = msg.split(',');
            const [xStr, yStr] = pos.split(':');

            enemyWorms.get(id).extend({ x: parseInt(xStr), y: parseInt(yStr) });

            break;
        }
    }
}

const init = () => {
    ws = new WebSocket("ws://localhost:8001");
    ws.onmessage = ({ data }) => {
        const unparsedWorms = data.split("\n");

        let positions = parseNewEvent(unparsedWorms[0])[1];

        playerWorm = new PlayerWorm(
            positions,
            generateRandomColour(),
            generateRandomColour(),
        );

        for(let i = 1; i < unparsedWorms.length; i++){
            const [id, positions] = parseNewEvent(unparsedWorms[i]);

            const enemyWorm = new EnemyWorm(
                positions,
                generateRandomColour(),
                generateRandomColour(),
            );

            enemyWorms.set(id, enemyWorm);
        }

        loading.style.visibility = "hidden";
        progressBox.style.visibility = "visible";

        ws.onmessage = handleWsMsg;
    };
    ws.onopen = () => ws.send("INIT");
}

init();
