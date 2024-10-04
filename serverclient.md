Client connects via HTTP request, is served HTML, CSS, JS.

Client then connects via WS. When the client establishes a connection it sends "INIT" to the server.

INIT:
    -Client initiates by sending "INIT"
    -Server then replies with message detailing positions of foods and other worms on server. Positions in the format x:y,x:y,.....

        INIT
        ID,NEWWORMPOSITIONS|EXISTINGWORMPOSITIONS(NEWLINE FOR EACH WORM, EACH WORM STARTS WITH THEIR ID FOLLOWED BY COMMA)|FOODPOSITIONS

    eg.

        INIT
        11,1:1,1:2,1:3|1,5:5,5:6,5:7,5:8
        3,8:1,8:2,9:2
        4,7:9,7:10,7:11|
        1:1,6:3,2:2,12:12

    -Server will then broadcast NEW message to other worms

NEW:
    -Client initiates by sending "INIT"
    -Broadcasted to all other clients except initiating client by server

        NEW
        ID,NEWWORMPOSITIONS

    eg.

        NEW
        8,1:1,1:2,1:3

MOVE:
    -Client initiates by sending "MOVE"

        MOVE
        DIRECTION(U,D,L,R)

    eg.

        MOVE
        U

    -Server updates state and then broadcasts move to other worms

        MOVE
        ID,DIRECTION(U,D,L,R)

    eg.

        MOVE
        3,R

    -If head cell contained food, server will broadcast CONSUMEFOOD message to other worms

CONSUMEFOOD:
    -Initiated when a worm moves to a cell containing food

        CONSUMEFOOD
        ID,POSITION

    eg.

        CONSUMEFOOD
        3,5:23

EXTEND:
    -Initiated when a worm consumes enough food to extend

        EXTEND
        ID,NEWPOSITION

    eg.

        EXTEND
        3,5:23
