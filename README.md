# Distributed Hash Computation

Authors: [Qiyuan Liang](https://github.com/IYuan505), [Qiyuan Dong](https://github.com/akaqyd), [Yanni Zhang](https://github.com/YanniZhangYZ)<br>
Overleaf Link: [Link](https://www.overleaf.com/project/639f1a97a9ac119914c1a0e3)

## Motivation
We want to crack a list of password hashes, they are salted. We want to distribute a portion of the list to one peer, the peer could compute the pre-image of the hash and return the result back to us.

In such a way, we could crack the list much faster. To allow the mechanism to work, we would like to have some incentives inside. For example, if some peers crack a hash, they will receive some coins inside our system. And to submit a task, the peer should spend some coins. Here, we utilize the help from Blockchain and smart contracts. However, naively distributing the list of hashes to random peers is not very efficient and suboptimal. We notice that dictionary attacks are not frequently seen nowadays because every hash is salted. If we could distribute the hashes according to salt and one peer knows that only some salt values can come to him, then, the corresponding peers could pre-compute a dictionary, which helps to crack the password. Here, we deploy distributed hash table to do this task efficently.
