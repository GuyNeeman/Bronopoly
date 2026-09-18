export default function StartScreen({ setScreen }) {
    return (
        <>
            <h1>Bronopoly</h1>

            <p>
                Buy properties, upgrade them, and collect rent.
                The last player with money wins.
            </p>

            <button onClick={() => setScreen("players")}>
                Create
            </button>

            <button onClick={() => setScreen("join")}>
                Join
            </button>

            <button onClick={() => setScreen("statistic")}>
                Statistic
            </button>

            <button onClick={() => setScreen("settings")}>
                Settings
            </button>
        </>
    );
}