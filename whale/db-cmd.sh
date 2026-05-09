if [ "$1" == "delete" ]; then
  sqlite3 app.db "delete from people"
else
  sqlite3 app.db "select * from people"
fi

