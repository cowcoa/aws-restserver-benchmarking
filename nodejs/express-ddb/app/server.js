'use strict';

// Depend npm modules
const moment = require('moment-timezone');
const express = require("express");
const bodyParser = require("body-parser");
const AWS = require('aws-sdk');
// Global objects, variables and init functions
const app = express();
const router = express.Router();
const dynamodb = new AWS.DynamoDB();
// Server port
const HTTP_PORT = 8387;
// DDB Info
const DDB_TABLE_NAME = process.env.DDB_TABLE_NAME
const DDB_GSI_NAME = process.env.DDB_GSI_NAME
const AWS_REGION = process.env.AWS_REGION

AWS.config.update({region:AWS_REGION});

app.use(bodyParser.urlencoded({ extended: false }));
app.use(bodyParser.json());

async function putChatInfo(name, comment, chatRoom) {
    console.log('Put chat info with name: ' + name + ', comment: ' + comment + ', chatRoom: ' + chatRoom);
    return new Promise((resolve, reject) => {
        let currentTime = moment().tz("Asia/Shanghai");
        let registerDate = currentTime.format();

        let params = {
            Item: {
                "Name": { S: name },
                "Time": { S: registerDate },
                "Comment": { S: comment },
                "ChatRoom": { S: chatRoom }
            },
            ConditionExpression: "attribute_not_exists(UserId)",
            ReturnConsumedCapacity: "TOTAL",
            TableName: DDB_TABLE_NAME
        };
        dynamodb.putItem(params, function(err, data) {
            if (err) {
                reject(err);
            } else {
                resolve();
            }
        });

    });
}

async function queryChatRecords(chatRoom) {
    console.log('Get chat info with chatRoom: ' + chatRoom);
    return new Promise((resolve, reject) => {
        let queryParams = {
            ExpressionAttributeValues: {
                ":c": { S: chatRoom }
            },
            KeyConditionExpression: "ChatRoom = :c",
            Limit: 10,
            TableName: DDB_TABLE_NAME,
            IndexName: DDB_GSI_NAME,
        };
        dynamodb.query(queryParams, function(err, data) {
            if (err) {
                reject(err);
            } else {
                console.log('response of queryChatRecords.query: ' + JSON.stringify(data));
                resolve(data);
            }
        });
    });
}

// Request pre-process function
router.use(function (req, res, next) {
    console.log('Time:', Date.now());
    console.log('DDB_TABLE_NAME:', DDB_TABLE_NAME);
    console.log('DDB_GSI_NAME:', DDB_GSI_NAME);
    console.log('AWS_REGION:', AWS_REGION);
    
    /*
    // Register disconnection callback
    req.on("close", function() {
        // request closed unexpectedly
        console.log('request closed unexpectedly');
    });
    req.on("end", function() {
        // request ended normally
        console.log('request ended normally');
    });
    */
    
    console.log('Received request: ' + req.method + ' '+ req.originalUrl + ' ' + JSON.stringify(req.body));
    next('route');
});

router.post('/put', async (req, res, next) => {
    if (req.get('Content-Type') != 'application/json') {
        console.log('Invalid Content-Type header: ' + req.get('Content-Type'));
        res.status(400).json({ error: 'Content-Type must be application/json' });
        return;
    }
    
    if (typeof req.body.name !== 'string' || 
        typeof req.body.comment !== 'string' || 
        typeof req.body.chatRoom !== 'string') {
        console.log('Invalid body');
        res.status(400).json({ error: 'Invalid Body: Missing required fields' });
        return;
    }

    try {
        await putChatInfo(req.body.name, req.body.comment, req.body.chatRoom);
    }
    catch (err) {
        console.error('Error: ' + err.message);
        if (err.statusCode == undefined) {
            err.statusCode = 500;
        }
        res.status(err.statusCode).json({ error: err.message });
    }

    res.status(201).end();
});

router.get('/get', async (req, res, next) => {
    if (typeof req.query.chatroom !== 'string') {
        console.log('Invalid query parameter');
        res.status(400).json({ error: 'Invalid Query Parameter: Missing required parameter chatroom' });
        return;
    }

    try {
        let result = await queryChatRecords(req.query.chatroom);
        let chatRecords = [];
        result.Items.forEach(element => {
            console.log('chatRecord: ' + JSON.stringify(element));
            let chatRecord = {
                name: element.Name.S,
                comment: element.Comment.S,
                time: element.Time.S
            }
            chatRecords.push(chatRecord);
        });
        res.status(200).json(chatRecords);
    }
    catch (err) {
        console.error('Error: ' + err.message);
        if (err.statusCode == undefined) {
            err.statusCode = 500;
        }
        res.status(err.statusCode).json({ error: err.message });
    }
});

// Default error handler
/*
app.use(function (err, req, res, next) {
    console.error('Internal Server Error');
    console.error(err.stack);
    res.status(500).send('Internal Server Error');
});

// Default response for any other request
app.use(function(err, req, res, next) {
    console.warn('Resource does not exist');
    res.status(404).end();
});
*/

// Start server
app.use('/', router);
app.listen(HTTP_PORT, () => {
    console.log("Server running on port %PORT%".replace("%PORT%",HTTP_PORT));
});
