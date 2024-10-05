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
    -Server initiates on set interval
    -If worm moves into food position, CONSUMEFOOD will be sent
    -Broadcasted to all clients

        MOVE
        ID,POSITIONS
        ID1,POSITIONS1....

    eg.

        MOVE
        1,2:2,2:3,2:4
        4,8:8,7:8,6:8,5:8

CONSUMEFOOD:
    -Initiated when a worm moves to a cell containing food

        CONSUMEFOOD
        ID,POSITION|FOODCONSUMED/FOODNEEDED

    eg.

        CONSUMEFOOD
        3,5:23|2/6

    -If worm has now eaten enough to extend, the server will add to their positions

CHANGEDIRECTION:
    -Initiated by client when changing direction
    -Updates worm's direction on server

        CHANGEDIRECTION
        ID,DIR

    eg.

        CHANGEDIRECTION
        3,R

DISCONNECT:
    -Broadcasted to all clients when a client disconnects

        DISCONNECT
        ID

    eg.

        DISCONNECT
        3

COLLIDE:
    -Sent to worm whose length is reduced in a collision

        COLLIDE
        NEWFOODCONSUMED/NEWFOODNEEDED

    eg.

        COLLIDE
        0/3
