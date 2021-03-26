package main

type Entry struct {
	Id      int         /* Entry index. Increases monotonically */
	Term    int         /*  Term when entry was received */
	Command interface{} /* Command value from client */
}
