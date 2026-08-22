import { useEffect } from "react";

export default function PlayerCreationScreen({ playerCount, players, setPlayers, next, back, startMoney }) {

    let Color = "blue"

    useEffect(() => {
        setPlayers(
            Array.from({ length: playerCount }, (_, i) => ({
                id: i,
                name: "Player " + (i+1),
                money: startMoney,
                color: Color,
                position: 0,
            }))
        );
    }, [playerCount, startMoney]);

    function updatePlayer(index, field, value) {
        const updated = [...players];
        updated[index][field] = value;
        setPlayers(updated);
    }

    return (
        <>
            <h2>Players</h2>

            {players.map((player, i) => (
                <div key={player.id} className="player-row">
                    <input
                        type="text"
                        value={player.name}
                        onChange={(e) => updatePlayer(i, "name", e.target.value)}
                    />y
                </div>
            ))}

            <div>
                <button onClick={back}>Back</button>
                <button onClick={next}>Next</button>
            </div>
        </>
    );
}