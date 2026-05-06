import { useEffect } from "react";

const COLORS = [
    "#ef4444", // red
    "#3b82f6", // blue
    "#22c55e", // green
    "#eab308", // yellow
    "#a855f7", // purple
    "#f97316", // orange
];

function PlayerCreationScreen({playerCount, players, setPlayers, next, back, startMoney}) {

    useEffect(() => {
        const newPlayers = Array.from({ length: playerCount }, (_, i) => ({
            id: i,
            name: `Player ${i + 1}`,
            color: COLORS[i],
            money: startMoney,
            position: 0,
        }));

        setPlayers(newPlayers);
    }, [playerCount, startMoney]);

    function updatePlayer(index, field, value) {
        const updated = [...players];
        updated[index][field] = value;
        setPlayers(updated);
    }

    return (
        <>
            <h2>Players</h2>
            <p>Set names and pick a color</p>

            {players.map((player, i) => (
                <div key={player.id} className="player-card">

                    <input
                        type="text"
                        value={player.name}
                        onChange={(e) =>
                            updatePlayer(i, "name", e.target.value)
                        }
                    />

                    <div
                        className="color-preview"
                        style={{ backgroundColor: player.color }}
                    />

                    <div className="color-palette">
                        {COLORS.map((color) => {
                            const isSelected = player.color === color;

                            const isTaken = players.some(
                                (p, index) =>
                                    p.color === color && index !== i
                            );

                            return (
                                <div
                                    key={color}
                                    className={`color-box 
                                        ${isSelected ? "selected" : ""} 
                                        ${isTaken ? "disabled" : ""}`}
                                    style={{ backgroundColor: color }}
                                    onClick={() => {
                                        if (!isTaken) {
                                            updatePlayer(i, "color", color);
                                        }
                                    }}
                                />
                            );
                        })}
                    </div>

                </div>
            ))}

            <div className="buttons">
                <button onClick={back}>Back</button>
                <button onClick={next}>Next</button>
            </div>
        </>
    );
}

export default PlayerCreationScreen;