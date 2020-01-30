const createError = require("http-errors");
const express = require("express");
const path = require("path");
const cookieParser = require("cookie-parser");
const logger = require("morgan");
const helmet = require("helmet");

const indexRouter = require("./routes/index");
const puzzlesRouter = require("./routes/puzzles");
const authRouter = require("./routes/auth");

const { isDatabaseSeeded } = require("./db/checkDatabases");
const loadSeedData = require("./db/loadSeedData");

const redis = require("redis");
const session = require("express-session");
const RedisStore = require("connect-redis")(session);

const client = redis.createClient({
    host: "redis",
});
const app = express();

app.use(helmet());
app.use(
    session({
        store: new RedisStore({ client }),
        secret: process.env.SESSION_SECRET,
        resave: false,
        secure: process.env.NODE_ENV === "production",
        cookie: {
            // 28 days
            maxAge: 1000 * 60 * 60 * 24 * 7 * 4,
        },
        name: "alice",
    })
);
app.use(logger("dev"));
app.use(express.json());
app.use(express.urlencoded({ extended: false }));
app.use(cookieParser());
app.use(express.static(path.join(__dirname, "public")));

// After boot, check if the database has been seeded. If not, seed it.
// Maybe the seeder should always run, but check the `source_id` of each
// inserted object to see if it exists before inserting it. The problem
// with that is that this file reloads every time a file is changed due
// to nodemon.
const { shouldSeedProjectEuler } = isDatabaseSeeded({ projectEuler: true });
if (shouldSeedProjectEuler) {
    console.log("loading seed data");
    loadSeedData();
} else {
    console.log("NOT loading seed data");
}

// Mount the routers
app.use("/", indexRouter);
app.use("/puzzles", puzzlesRouter);
app.use("/auth", authRouter);

// A temporary middleware function to see what URL is being hit.
// app.use(function(req, res, next) {
//     console.log("URL", req.originalUrl);
//     next();
// });


// catch 404 and forward to error handler
app.use(function(req, res, next) {
    next(createError(404));
});

// error handler
app.use(function(err, req, res, next) {
    // set locals, only providing error in development
    res.locals.message = err.message;
    res.locals.error = req.app.get("env") === "development" ? err : {};

    // render the error page
    res.status(err.status || 500);
    res.json({ error: "error" });
});

module.exports = app;
