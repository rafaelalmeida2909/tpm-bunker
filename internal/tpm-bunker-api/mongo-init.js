// mongo-init.js
db = db.getSiblingDB("tpmbunker_db");
db.createCollection("tpmbunker_data");
db.createUser({
  user: "tpmbunker_api_user",
  pwd: "tpmbunker_api_password",
  roles: [
    {
      role: "readWrite",
      db: "tpmbunker_db",
    },
  ],
});
