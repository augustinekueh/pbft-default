# pbft-default
 A project to study about pBFT algorithms; starting with the original.

Steps:
1) Run the batch files : client.bat and node.bat to generate the associated folder and keys, then close all the cmds.
2) Run the batch files again to see the outcome.

Important:
1) Make sure to edit the main.go at line 33 to change the .json file based on the number of nodes chose in Step 1 (the node batch file; node_(number of nodes).bat).
2) After saving the changes, the .exe file has to be rebuild with the command 'go build' (require to install Go package in your machine).
3) Currently, node_128.json cannot run yet.(fixed)
