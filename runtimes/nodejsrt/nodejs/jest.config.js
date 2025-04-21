// eslint-disable-next-line no-undef
module.exports = {
    preset: "ts-jest",
    testEnvironment: "node",
    transform: {
        "^.+\\.ts$": [
            "ts-jest",
            {
                tsconfig: "./tsconfig.strict.json"
            }
        ]
    },
    testPathIgnorePatterns: ["/node_modules/", "/build/",
        "/dist/",
        "/examples-build/",
        "/testdata/",
        "/dev/"],
    watchPathIgnorePatterns: ["/node_modules/"]
};
