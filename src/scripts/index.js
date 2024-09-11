const grid = document.getElementById("wormo-grid");
const foodCounter = document.getElementById("food-counter");
const foodNeeded = document.getElementById("food-needed");
const progressBar = document.getElementsByClassName("progress-inner")[0];

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

const removeFoodFromCell = ({x, y}) => {
    const index = x + y * GRID_COLS;
    console.debug(`Removing worm food from ${x},${y}`)

    const cell = grid.children.item(index);
    cell.innerHTML = "";
    cell.classList.remove("worm-food");
    cell.style.removeProperty("colour");

    wormFoodLocations.delete(`${x}-${y}`);
};

const generateRandomColour = () => '#' + (Math.random().toString(16) + "000000").substring(2,8);

class Worm {
    /**
        @positions {x: number, y: number}[] tail->head
        @colour string
        @headColour string
    **/
    constructor(positions, colour, headColour) {
        this.positions = positions;
        this.colour = colour;
        this.headColour = headColour;
        this.foodConsumed = 0;

        for(let i = 0; i < positions.length - 1; i++){
            addColourToCell(positions[i], colour);
        }

        addColourToCell(positions[positions.length - 1], headColour);
    }

    /**
        @direction "U" | "D" | "L" | "R"
    **/
    move(direction) {
        let headPos = this.positions[this.positions.length - 1];
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

        for(let i = 0; i < this.positions.length - 1; i++){
            let pos = this.positions[i];
            removeColourFromCell(pos);

            const nextPos = this.positions[i + 1];

            pos.x = nextPos.x;
            pos.y = nextPos.y;
        }

        for(let i = 0; i < this.positions.length - 1; i++){
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
        const tailPos = this.positions[0];
        const newPos = findAdjacentFreeSpace(tailPos);

        const newPositions = [
            newPos,
            ...this.positions,
        ];

        this.positions = newPositions;
        this.foodConsumed = 0;
    }
}

let playerWorm = new Worm(
    [
        {x: 0, y: 1},
        {x: 1, y: 1},
        {x: 1, y: 2},
    ],
    generateRandomColour(),
    generateRandomColour(),
);

addEventListener("keydown", (event) => {
    if(event.repeat){
        return;
    }

    if(event.key === "ArrowUp"){
        playerWorm.move('U');
    } else if(event.key === "ArrowDown"){
        playerWorm.move('D');
    } else if(event.key === "ArrowLeft"){
        playerWorm.move('L');
    } else if(event.key === "ArrowRight"){
        playerWorm.move('R');
    }
});

const spawnFood = () => {
    const multiplier = Math.floor(Math.random() * 3);
    console.debug("Multiplier " + multiplier);

    for(let i = 0; i < TOTAL_SPACES / 50 * multiplier; i++){
        const randX = Math.floor(Math.random() * GRID_COLS);
        const randY = Math.floor(Math.random() * GRID_ROWS);

        const randPos = {x: randX, y: randY};

        if(checkCellHasFood(randPos)){
            continue;
        }

        const colour = generateRandomColour();
        addFoodToCell(randPos, colour);
    }
};

spawnFood();
setInterval(spawnFood, 5000);
