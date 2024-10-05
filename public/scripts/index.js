const grid = document.getElementById("wormo-grid");
const foodCounter = document.getElementById("food-counter");
const foodNeeded = document.getElementById("food-needed");
const progressBar = document.getElementById("progress-inner");
const loading = document.getElementById("ui-loading");
const progressBox = document.getElementById("ui-progress");

const TOTAL_SPACES = GRID_ROWS * GRID_ROWS;

const generateRandomColour = () => '#' + (Math.random().toString(16) + "000000").substring(2,8);

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
};

const removeFoodFromCell = ({x, y}) => {
    const index = x + y * GRID_COLS;
    console.debug(`Removing worm food from ${x},${y}`)

    const cell = grid.children.item(index);
    cell.innerHTML = "";
    cell.classList.remove("worm-food");
    cell.style.removeProperty("colour");
};


const updateFoodCounter = (consumed, needed) => {
    foodCounter.innerHTML = consumed;
    foodNeeded.innerHTML = needed;

    const percentage = consumed / needed * 100;
    progressBar.style.width = percentage + '%';
};

const parsePosition = (str) => {
    const [x, y] = str.split(':');
    return {x: parseInt(x), y: parseInt(y)};
};

const parsePositions = (str) => {
    const splitStr = str.split(',');
    const positions = Array(splitStr.length);

    for(let i = 0; i < splitStr.length; i++){
        positions[i] = parsePosition(splitStr[i]);
    }

    return positions;
};

const parseNewEvent = (str) => {
    const firstComma = str.indexOf(',');
    const id = str.slice(0, firstComma);
    const positions = parsePositions(str.slice(firstComma + 1));

    return [id, positions];
};

class Worm {
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
        @positions {x: number, y: number}[] head->tail
    **/
    updatePositions(positions) {
        this.clearPositions();

        addColourToCell(positions[0], this.headColour);

        for(let i = 1; i < positions.length; i++){
            addColourToCell(positions[i], this.colour);
        }

        this.positions = positions;
    }

    clearPositions() {
        for(const position of this.positions){
            removeColourFromCell(position);
        }
    }
}

let worms = new Map();
let ws;
let isInitialised = false;

const wsEvents = {
    MOVE: "MOVE",
    SPAWNFOOD: "SPAWNFOOD",
    CONSUMEFOOD: "CONSUMEFOOD",
    INIT: "INIT",
    CHANGEDIR: "CHANGEDIR",
    NEW: "NEW",
    DISCONNECT: "DISCONNECT",
    COLLIDE: "COLLIDE",
};

const handleWsMsg = ({ data }) => {
    console.debug("ws msg: " + data);

    const firstNewLine = data.indexOf("\n");

    const event = data.slice(0, firstNewLine);
    const msg = data.slice(firstNewLine + 1, data.length);

    if(event != wsEvents.INIT && !isInitialised){
        return;
    }

    switch(event){
        case wsEvents.CONSUMEFOOD: {
            const [idPosMsg, consumedNeededMsg] = msg.split('|');
            const [id, unparsedPosition] = idPosMsg.split(',');

            const position = parsePosition(unparsedPosition);
            removeFoodFromCell(position);

            if(id === playerId){
                const [consumed, needed] = consumedNeededMsg.split('/');
                updateFoodCounter(consumed, needed);
            }

            break;
        }
        case wsEvents.MOVE: {
            for(const unparsedWorm of msg.split("\n")){
                const [id, positions] = parseNewEvent(unparsedWorm);

                worms.get(id).updatePositions(positions);
            }

            break;
        }
        case wsEvents.SPAWNFOOD: {
            const foodPositions = parsePositions(msg);

            for(const foodPosition of foodPositions){
                addFoodToCell(foodPosition, generateRandomColour());
            }

            break;
        }
        case wsEvents.COLLIDE: {
            const [consumed, needed] = msg.split('/');
            updateFoodCounter(consumed, needed);

            break;
        }
        case wsEvents.DISCONNECT: {
            worms.get(msg).clearPositions();
            worms.delete(msg);

            break;
        }
        case wsEvents.NEW: {
            const [id, positions] = parseNewEvent(msg);

            const worm = new Worm(
                positions,
                generateRandomColour(),
                generateRandomColour(),
            );

            worms.set(id, worm);

            break;
        }
        case wsEvents.INIT: {
            let [playerWormMsg, enemyWormsMsg, foodMsg] = msg.split('|');

            let [id, positions] = parseNewEvent(playerWormMsg);

            const playerWorm = new Worm(
                positions,
                generateRandomColour(),
                generateRandomColour(),
            );

            playerId = id;

            worms.set(playerId, playerWorm);

            if(enemyWormsMsg != ""){
                const unparsedEnemyWorms = enemyWormsMsg.split("\n");

                for(const unparsedEnemyWorm of unparsedEnemyWorms){
                    let [id, positions] = parseNewEvent(unparsedEnemyWorm);

                    const enemyWorm = new Worm(
                        positions,
                        generateRandomColour(),
                        generateRandomColour(),
                    );

                    worms.set(id, enemyWorm);
                }
            }

            if(foodMsg !== ""){
                const foodPositions = parsePositions(foodMsg);

                for(const foodPosition of foodPositions){
                    addFoodToCell(foodPosition, generateRandomColour());
                }
            }

            loading.style.visibility = "hidden";
            progressBox.style.visibility = "visible";

            isInitialised = true;

            break;
        }
    }
}

const init = () => {
    ws = new WebSocket(document.URL.replace("http", "ws").replace("8000", "8001"));
    ws.onmessage = handleWsMsg;
    ws.onopen = () => {
        ws.send("INIT");

        addEventListener("keydown", ({ key, repeat }) => {
            if(repeat){
                return;
            }

            let dir = null;

            switch(key){
                case "ArrowUp":
                    dir = 'U';
                    break;
                case "ArrowDown":
                    dir = 'D';
                    break;
                case "ArrowLeft":
                    dir = 'L';
                    break;
                case "ArrowRight":
                    dir = 'R';
                    break;
                default:
                    return;
            }

            ws.send(wsEvents.CHANGEDIR + "\n" + dir);
        });
    };
};

init();
