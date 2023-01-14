# Distributed Hash Computation

Authors: [Qiyuan Liang](https://github.com/IYuan505), [Qiyuan Dong](https://github.com/akaqyd), [Yanni Zhang](https://github.com/YanniZhangYZ)<br>
Overleaf Link: [Link](https://www.overleaf.com/project/639f1a97a9ac119914c1a0e3)

## To run the program

In the root directory, there is a `main.go` file, which provides an entry point for the command line interface.

Run `go run .` to initiate a single node instance.

The node is assigned a random UDP address for communication purposes. You are expected to run multiple instances of nodes, e.g., at different command line tabs. After running `go run .`, the node's information is displayed on the screen, with the UDP address, and its own Chord ID. To allow users to submit tasks to crack passwords, you should at least join a Chord ring. You should specify one peer address that is inside the Chord ring. After you have joined the Chord ring, you could submit tasks, and receive the task results. Please follow the command line instructions. When you are done, you could either leave Chord, which will bring you back to the initial user interface, or exit directly.

## Motivation

We want to crack a list of password hashes, they are salted. We want to distribute a portion of the list to one peer, the peer could compute the pre-image of the hash and return the result back to us.

In such a way, we could crack the list much faster. To allow the mechanism to work, we would like to have some incentives inside. For example, if some peers crack a hash, they will receive some coins inside our system. And to submit a task, the peer should spend some coins. Here, we utilize the help from [Blockchain](https://en.wikipedia.org/wiki/Blockchain) and [smart contracts](https://en.wikipedia.org/wiki/Smart_contract). However, naively distributing the list of hashes to random peers is not very efficient and suboptimal. We notice that dictionary attacks are not frequently seen nowadays because every hash is salted. If we could distribute the hashes according to salt and one peer knows that only some salt values can come to him, then, the corresponding peers could pre-compute a dictionary, which helps to crack the password. Here, we deploy distributed hash table to do this task efficiently. We use [Chord](https://en.wikipedia.org/wiki/Chord_(peer-to-peer)).

## Design

### Password Cracker

The implementation could be found at `/peer/impl/password_cracker`.

The password cracker is the main application that we support. It supports two main APIs to users, allowing users to submit a task, and receive the result of a task.

1. When a user submits a task, he will first contact Chord, to find out which peer he should go to and ask for the result. Then, he submits the task to the peer.
2. When a user receives the result of a task, he could receive nothing, which means the password is not successfully cracked. Or he successfully receives the result, i.e., the pre-image of the hash.

When cracking a password locally, the password cracker uses a dictionary attack. It will pre-compute the hashes of all words inside the dictionary, and stores them. Exactly how many dictionaries a node needs to pre-compute depends on the DHT. A node in Chord is responsible for the key range from its predecessor and its own ID. Therefore, a node computes dictionaries with salt values range $(predecessor ID, its ID]$. The predecessor of a node may change when a new node joins the Chord, or an old node leaves the Chord.

### Chord

The implementation could be found at `/peer/impl/chord`.

The Chord is responsible for the efficient lookup of a key value. In our case, the key would be the salt of the hash. For each Chord node, it is responsible for the key range $(predecessor ID, its ID]$. For example, the Chord node A's ID is 129, and it has a predecessor with ID 22. Then it is responsible for the key range $(22, 129]$. If a node inside the Chord ring looks up key = 100, it will end up finding Chord node A. Once Chord receives some update regarding the key range it is responsible for, it will notify Password Cracker of the update. The update comes from two scenarios, when some nodes join the Chord ring, or some nodes leave the Chord ring. For our implementation, we do not consider the case of node failure (silent leave). Therefore, Chord also supports two main APIs.

1. Chord Join: it allows a Chord node to join an existing Chord ring, it includes finding out the correct successor of the node. The predecessor and finger table entry is updated by regular running daemons.
2. Chord Leave: it allows a Chord node to leave an existing Chord ring. This includes notifying our predecessor, i.e., our predecessor should use our successor as the new successor, and notifying our successor, i.e., his predecessor is no longer valid.

### Blockchain

### Smart Contract

The implementation could be found at `/peer/impl/contract`.

In this project, we created a simplified version of Ethereum. Our own smart contract is based on the basic concepts of Ethereum but has been scaled down specifically for this project. We customized a set of primitives for the contract code, which allows us to have more control and flexibility in the implementation of our smart contract.

The smart contract involves two components. The first, the lexer and parser, converts user-provided plain text into code tokens, verifies syntax, and constructs AST. The second, the interpreter, traverses the AST, evaluates conditions, and gathers the actions specified in the smart contract. The blockchain subsequently uses the interpreter's output to execute the contract. Smart contract supports two main APIs that will be used in blockchain.

1. Check Assumption: this function assesses if the assumption criteria are satisfied before executing the contract. In our case, we verify if the smart account has sufficient balance for the task finisher's reward. If the check fails, the contract will not proceed with the execution.

2. Gather Actions: after passing the assumption check, this function deals with the IF-THEN clause in the contract. It first evaluates if the condition specified by the IF statement is met. In our case, the task finisher will verify the correctness of the task by recalculating the hash of the cracked password with the associated salt. If the condition is satisfied, the actions outlined in the contract are gathered and submitted to the blockchain for execution.
