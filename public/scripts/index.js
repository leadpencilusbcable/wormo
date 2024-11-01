const grid = document.getElementById("wormo-grid");
const foodCounter = document.getElementById("food-counter");
const foodNeeded = document.getElementById("food-needed");
const progressBar = document.getElementById("progress-inner");
const loading = document.getElementById("ui-loading");
const progressBox = document.getElementById("ui-progress");

let bombImageSrc;

(async () => {
    const res = await fetch(document.baseURI + "/images/bomb.png");

    if(res.status === 200){
        bombImageSrc = URL.createObjectURL(await res.blob());

        if(bombs){
            for(const [_, bomb] of bombs){
                bomb.bombImage.src = bombImageSrc;
            }
        }
    } else{
        console.error("Unable to fetch bomb image");
    }
})();

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

const addBorderColourToCell = ({x, y}, colour) => {
    const index = x + y * GRID_COLS;
    console.debug(`Adding border colour ${colour} to ${x},${y}`)

    const cell = grid.children.item(index);
    cell.style["border-color"] = colour;
};

const removeBorderColourFromCell = ({x, y}) => {
    const index = x + y * GRID_COLS;
    console.debug(`Removing border colour from ${x},${y}`)

    const cell = grid.children.item(index);
    cell.style.removeProperty("border-color");
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

const parseSpawnBombEvent = (str) => {
    const [id, timeToDetonateSeconds, unparsedBombPosition, unparsedPositions] = str.split('|');

    return [id, parseInt(timeToDetonateSeconds), parsePosition(unparsedBombPosition), parsePositions(unparsedPositions)];
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

        for(let i = 1; i < positions.length; i++){
            addColourToCell(positions[i], this.colour);
        }

        addColourToCell(positions[0], this.headColour);

        this.positions = positions;
    }

    clearPositions() {
        for(const position of this.positions){
            removeColourFromCell(position);
        }
    }
}

class Bomb {
    /**
        @param {{x: number, y: number}} bombPosition The position of the bomb
        @param {{x: number, y: number}[]} positions The surrounding positions in the bomb's range
        @param {number} timeToDetonateSeconds How long for bomb to detonate
    **/
    constructor(bombPosition, positions, timeToDetonateSeconds) {
        this.bombPosition = bombPosition;
        this.positions = positions;
        this.timeToDetonateSeconds = timeToDetonateSeconds;

        const bombOverlay = document.createElement("div");
        bombOverlay.className = "bomb-overlay";
        bombOverlay.style.gridColumn = (positions[0].x + 1) + '/' + (positions[positions.length - 1].x + 2);
        bombOverlay.style.gridRow = (positions[0].y + 1) + '/' + (positions[positions.length - 1].y + 2);

        const bombTimer = document.createElement("span");
        bombTimer.className = "bomb-timer";
        bombTimer.innerHTML = timeToDetonateSeconds;

        bombOverlay.appendChild(bombTimer);

        grid.appendChild(bombOverlay);
        this.bombOverlay = bombOverlay;

        const bombImage = document.createElement("img");
        bombImage.className = "bomb-image";
        bombImage.src = bombImageSrc;
        bombImage.style.gridColumn = bombPosition.x + 1;
        bombImage.style.gridRow = bombPosition.y + 1;

        grid.appendChild(bombImage);
        this.bombImage = bombImage;

        this.intervalId = setInterval(this.decrement.bind(this), 1000);
    }

    decrement() {
        this.timeToDetonateSeconds--;
        this.bombOverlay.children[0].innerHTML = this.timeToDetonateSeconds;

        if(this.timeToDetonateSeconds <= 0){
            clearInterval(this.intervalId);
        }
    }

    detonate() {
        this.bombOverlay.remove();
        this.bombImage.remove();
    }
}

let worms = new Map();
let bombs = new Map();
let ws;
let isInitialised = false;

const wsEvents = {
    MOVE: "MOVE",
    SPAWNFOOD: "SPAWNFOOD",
    CONSUMEFOOD: "CONSUMEFOOD",
    SPAWNBOMB: "SPAWNBOMB",
    DETONATEBOMB: "DETBOMB",
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

                if(id === playerId && positions.length !== worms.get(playerId).positions.length){
                    updateFoodCounter(0, positions.length * LEVEL_MULTIPLIER)
                }

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
        case wsEvents.SPAWNBOMB: {
            const [id, timeToDetonateSeconds, bombPosition, positions] = parseSpawnBombEvent(msg);

            const bomb = new Bomb(bombPosition, positions, timeToDetonateSeconds);
            bombs.set(id, bomb);

            break;
        }
        case wsEvents.DETONATEBOMB: {
            const [bombId, unparsedWorms] = msg.split('|');

            bombs.get(bombId).detonate();
            bombs.delete(bombId);

            if(unparsedWorms !== undefined){
                for(const unparsedWorm of unparsedWorms.split("\n")){
                    const [id, positions] = parseNewEvent(unparsedWorm);

                    worms.get(id).updatePositions(positions);
                }
            }

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
            let [playerWormMsg, enemyWormsMsg, foodMsg, bombMsg] = msg.split('|');

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

            if(bombMsg !== ""){
                const unparsedBombs = bombMsg.split("\n");

                for(const unparsedBomb of unparsedBombs){
                    const bombData = unparsedBomb.split(',');

                    const id = bombData[0];
                    const timeToDetonateSeconds = bombData[1];
                    const unparsedBombPosition = bombData[2];
                    const unparsedPositions = bombData.slice(3);

                    const positions = [];

                    for(const unparsedPosition of unparsedPositions){
                        positions.push(parsePosition(unparsedPosition));
                    }

                    const bomb = new Bomb(parsePosition(unparsedBombPosition), positions, parseInt(timeToDetonateSeconds));
                    bombs.set(id, bomb);
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
    console.log("%cWelcome to\n%cW%cO%cR%cM%cO",
        "font-size: 20px",
        "font-size: 50px; color: " + generateRandomColour(),
        "font-size: 50px; color: " + generateRandomColour(),
        "font-size: 50px; color: " + generateRandomColour(),
        "font-size: 50px; color: " + generateRandomColour(),
        "font-size: 50px; color: " + generateRandomColour(),
    );

    const wsUrl = new URL(document.URL);
    wsUrl.port = WS_PORT;

    ws = new WebSocket(wsUrl);
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
