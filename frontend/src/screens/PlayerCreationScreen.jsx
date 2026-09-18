import { useEffect } from "react";

const COLORS = ["red", "blue", "green", "yellow", "purple"];

export default function PlayerCreationScreen({ playerCount, players, setPlayers, next, back, startMoney }) {
    useEffect(() => {
        setPlayers(
            Array.from({ length: playerCount }, (_, i) => ({
                name: `Player ${i + 1}`,
                money: startMoney,
                color: COLORS[i % COLORS.length],
                position: 0,
            }))
        );
    }, [playerCount, startMoney]);

    function updatePlayer(index, field, value) {
        setPlayers(players.map((player, i) => (i === index ? { ...player, [field]: value } : player)));
    }

    if (!players) return null;

    return (
        <>
            <h2>Players</h2>

            {players.map((player, index) => (
                <div className="player-row" key={index}>
                    <input
                        type="text"
                        value={player.name}
                        onChange={(e) => updatePlayer(index, "name", e.target.value)}
                    />
                    <select
                        value={player.color}
                        onChange={(e) => updatePlayer(index, "color", e.target.value)}
                    >
                        {COLORS.map((color) => (
                            <option key={color} value={color}>
                                {color}
                            </option>
                        ))}
                    </select>
                </div>
            ))}

            <div>
                <button onClick={back}>Back</button>
                <button onClick={next}>Next</button>
            </div>
        </>
    );
}
