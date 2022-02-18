db.auth('mongodb', 'mongodb')
db = db.getSiblingDB('dips')
db.createUser({
    user: "dips",
    pwd: "dips",
    roles: [{
        role: "readWrite",
        db: "dips"
    }]
});