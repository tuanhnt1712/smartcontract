pragma solidity ^0.4.19;

/**
 * The Anime contract does this and that...
 */
contract Anime {
    address public owner;
    mapping (address => uint) public treasures;
  
    function Anime () {
        owner = msg.sender;
        // otakuToken = 1000000;
        treasures[owner] = 100;
    }

    function register(address user, uint amount) public {
        if (msg.sender != user) return;
        treasures[user] = amount;
    }
    

    function getTreasure(address _a) public constant returns(uint) {
        if (treasures[_a] != 0)
          return treasures[_a];
        return 0;
    }

    function addTreasure(address _a ,uint amount) public {
        if (msg.sender != owner) return;
        if (amount > 10)
          treasures[_a] += amount;
        else
          return;
    }

    function sendTreasure(address receiver, uint amount) public {
        if (msg.sender != owner) return;
        if (amount < 0) return;
        if (treasures[receiver] != 0) return;
        treasures[receiver] += amount;
        treasures[owner] -= amount;
    }
}
