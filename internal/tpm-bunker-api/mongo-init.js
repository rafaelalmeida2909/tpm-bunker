db.createUser({
    user: 'tpmbunker_api_user',
    pwd: 'tpmbunker_api_password',
    roles: [
      {
        role: 'readWrite',
        db: 'tpmbunker_db'
      },
      {
        role: 'dbAdmin',
        db: 'tpmbunker_db'
      }
    ]
  });
  
  db = db.getSiblingDB('tpmbunker_db');