Scalable chat application server using MySQL, WebSocket and Go 

Task to be done (Month1)
  1. Mysql connection pool (reusability) done
  2. CRUD operation  (reusability) done
  3. JWT authentication + third party authentication like facebook and google  (reusability) work in progress
  4. Redis implementation  (reusability)
  5. Mongodb and Cassandra   (reusability)
  6. API testing point dummy (no reusability)
  7. https server (reusability)
  8. web sockets  (done)
  8. webrtc  (done)
  9. adapter pattern (reusability)
  10. jenkins pipeline
  11. Task to be done (Month2)

  12. room based chat done (normal + emoji message done)
  13. webrtc based video and audio boardcasting room (done)
  14. ui client based on react done
  15. complex html message need to be implemented ( in progress)
  16. animated images and other kind of stuff need to be implemented (sharing picture and other stuff)
  17. screensharing need to be implemented
  18. optimization and performance need to be implemented 
DATABASE_KEY="root:manish@tcp(127.0.0.1:3306)/chatmsg"

project uses:-
----------------------------------------------------------------------------------------------------------
Gorilla Mux for rest 
Pion for webrtc (SFU)
mysql as database
jwt for auth 
react for frontend and react routing for moving between pages
project is in intial state right now so its not scalble for real production env
not stable but workable in most of the cases
----------------------------------------------------------------------------------------------------------

How to use this
--------------------------------
git clone https://github.com/manishchauhan/dugguGo.git
also import chatmsg.mwb to your mysql database
all connection setting in env file
go run main.go 
for ui you need to clone https://github.com/manishchauhan/duggugoui/tree/chat which contain ui stuff for this project
