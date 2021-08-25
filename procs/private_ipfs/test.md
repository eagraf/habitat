### Running / Testing
1. ./start.sh ipfs1 --> exit the editor
2. ./start.sh ipfs2 --> change ports to API: 5002, Gateway: 8081, TCP: 4002
3. 	cd
4. ./ipfs-swarm-key-gen > ipfs1/swarm.key
5. IPFS_PATH=ipfs1/ ipfs id
6. copy "ID" field = PEERID
7. ./bootstrap_peer.sh ipfs2/ 127.0.0.1 4001 12D3KooWBMUwzLnrH8DLrn5fWp13amWw2ufVZ6Qgg411vnZPrbWP PEERID
8. ./connect.sh ipfs1
9. ./connect.sh ipfs2
10. Make a file called 'hi.txt' with the contents 'Hello World!'
11. IPFS_PATH=ipfs1/ ipfs add hi.txt
12. IPFS_PATH=ipfs2/ ipfs cat QmfM2r8seH2GiRaC4esTjeraXEachRt8ZsSeGaWTPLyMoG
13. Should output Hello World!