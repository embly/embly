extern crate embly;

use {
    bcrypt::{hash, verify},
    bincode,
    cookie::Cookie,
    embly::{
        http::{Body, Request, ResponseWriter},
        kv,
        prelude::*,
        Error,
    },
    http::Method,
    regex::Regex,
    serde::{Deserialize, Serialize},
    serde_json::{json, Value},
    uuid::Uuid,
};

#[derive(Serialize, Deserialize, PartialEq, Debug)]
struct User {
    id: Vec<u8>,
    username: String,
    email: String,
    password_hash: String,
}

fn prepare_key(data: &[u8], prefix: u8) -> Vec<u8> {
    let mut key: Vec<u8> = vec![0; 1 + data.len()];
    key[0] = prefix;
    key[1..].copy_from_slice(&data);
    key
}

impl User {
    fn new() -> Self {
        Self {
            id: Vec::new(),
            username: String::new(),
            email: String::new(),
            password_hash: String::new(),
        }
    }
    fn new_from_bytes(data: &[u8]) -> Option<Self> {
        match bincode::deserialize(data) {
            Ok(session) => Some(session),
            Err(_) => None,
        }
    }
    async fn save(&self) -> Result<(), Error> {
        let set1 = kv::set(&prepare_key(&self.id, USER_KV_PREFIX), &self.to_bytes());
        let set2 = kv::set(
            &prepare_key(&self.email.as_bytes(), EMAIL_KV_PREFIX),
            &self.id,
        );
        let set3 = kv::set(
            &prepare_key(&self.username.as_bytes(), USERNAME_KV_PREFIX),
            &self.id,
        );
        set1.await?;
        set2.await?;
        set3.await?;
        Ok(())
    }
    async fn find(key: &[u8]) -> Option<Self> {
        if let Ok(user) = kv::get(&prepare_key(key, USER_KV_PREFIX)).await {
            Self::new_from_bytes(&user)
        } else {
            None
        }
    }
    fn to_bytes(&self) -> Vec<u8> {
        bincode::serialize(&self).unwrap()
    }
}

#[derive(Serialize, Deserialize, PartialEq, Debug)]
struct Session {
    token: Vec<u8>,
    user_id: Vec<u8>,
}

const USER_KV_PREFIX: u8 = 1;
const SESSION_KV_PREFIX: u8 = 2;
const EMAIL_KV_PREFIX: u8 = 3;
const USERNAME_KV_PREFIX: u8 = 4;

impl Session {
    fn new() -> Self {
        Self {
            token: rand::random::<[u8; 16]>().to_vec(),
            user_id: Vec::new(),
        }
    }
    fn new_from_bytes(data: &[u8]) -> Option<Self> {
        match bincode::deserialize(data) {
            Ok(session) => Some(session),
            Err(_) => None,
        }
    }
    async fn find(token: &[u8]) -> Option<Self> {
        if let Ok(session) = kv::get(&prepare_key(token, SESSION_KV_PREFIX)).await {
            Self::new_from_bytes(&session)
        } else {
            None
        }
    }
    async fn save(&self) -> Result<(), Error> {
        kv::set(
            &prepare_key(&self.token, SESSION_KV_PREFIX),
            &self.to_bytes(),
        )
        .await?;
        Ok(())
    }
    fn to_bytes(&self) -> Vec<u8> {
        bincode::serialize(&self).unwrap()
    }
}

fn write_validation_error(w: &mut ResponseWriter, msg: &str, field: &str) -> Result<(), Error> {
    println!("writing validation error {}", msg);
    let v = json!({ "msg": msg, "field": field });
    w.status(422)?;
    w.write(&serde_json::to_vec(&v)?)?;
    Ok(())
}

// curl -d '{"username":"max", "email":"max.t.mcdonnell@gmail.com", "password":"password"}' -H "Content-Type: application/json" -X POST http://localhost:8082/auth/register

fn is_palindrome(input: &str) -> bool {
    let s = input
        .chars()
        .filter(|&c| c.is_alphanumeric())
        .collect::<String>();
    s == s.chars().rev().collect::<String>()
}

async fn login(w: &mut ResponseWriter, user: User) -> Result<(), Error> {
    let mut session = Session::new();
    session.user_id = user.id.clone();
    let uuid_string = Uuid::from_slice(&session.token.clone())
        .unwrap()
        .to_hyphenated()
        .to_string();
    w.header("Set-Cookie", format!("_embly-session={}", uuid_string))?;
    session.save().await?;
    Ok(())
}

fn signout(w: &mut ResponseWriter) -> Result<(), Error> {
    w.write(
        json!({
            "success": true,
        })
        .to_string()
        .as_bytes(),
    )?;
    w.header("Set-Cookie", "_embly-session= ")
}

async fn current_user(req: Request<Body>) -> Result<Option<User>, Error> {
    let cookie = req.headers().get("Cookie");
    if cookie.is_none() {
        return Ok(None);
    }

    let cookiestr = cookie.unwrap().to_str()?;
    for cookie in cookiestr.split(";") {
        // ignore parse errors
        if let Ok(parsed) = Cookie::parse(cookie) {
            if parsed.name() == "_embly-session" {
                if let Ok(parsed_uuid) = Uuid::parse_str(parsed.value()) {
                    if let Some(session) = Session::find(parsed_uuid.as_bytes()).await {
                        if let Some(user) = User::find(&session.user_id).await {
                            return Ok(Some(user));
                        } else {
                            println!("can't find user");
                        }
                    } else {
                        println!("can't find session");
                    }
                } else {
                    println!("can't parse uuid");
                }
            }
        } else {
            println!("can't parse cookie");
        }
    }
    Ok(None)
}

async fn execute(mut req: Request<Body>, w: &mut ResponseWriter) -> Result<(), Error> {
    w.header("Access-Control-Allow-Origin", "*")?;
    if req.method() == Method::OPTIONS {
        w.status(200)?;
        w.header("Access-Control-Allow-Methods", "POST, GET, OPTIONS")?;
        w.header("Access-Control-Allow-Headers", "Content-Type")?;
        w.header("Access-Control-Allow-Credentials", "true")?;
        return Ok(());
    }

    if req.method() == Method::GET && req.uri().path().starts_with("/api/auth/user") {
        if let Ok(maybe_user) = current_user(req).await {
            if let Some(user) = maybe_user {
                let v = json!({ "username": user.username, "email": user.email });
                w.status(200)?;
                w.write(&serde_json::to_vec(&v)?)?;
                return Ok(());
            }
        }
        w.status(401)?;
        let v = json!({ "msg": "not allowed" });
        w.write(&serde_json::to_vec(&v)?)?;
        return Ok(());
    }
    if req.method() == Method::GET && req.uri().path().starts_with("/api/auth/sign-out") {
        signout(w)?;
        return Ok(());
    }

    if req.method() != Method::POST {
        w.status(405)?;
        w.write(b"method not allowed")?;
        return Ok(());
    }
    let body = match req.body_mut().bytes() {
        Ok(body) => body,
        Err(err) => {
            println!("{:?}", err);
            return write_validation_error(w, "error reading http body", "");
        }
    };
    println!("{:?}", ::std::str::from_utf8(&body).unwrap());
    let v: Value = serde_json::from_slice(&body)?;

    if req.uri().path().starts_with("/api/auth/register") {
        if v["email"].is_null() {
            return write_validation_error(w, "email is required", "email");
        }
        if v["username"].is_null() {
            return write_validation_error(w, "username is required", "username");
        }
        if v["password"].is_null() {
            return write_validation_error(w, "password is required", "password");
        }

        let email = v["email"].as_str().unwrap();
        if !email.contains("@") {
            return write_validation_error(w, "email address must be valid", "email");
        }

        let username = v["username"].as_str().unwrap().to_lowercase();
        if username.len() < 3 {
            return write_validation_error(
                w,
                "username must be 3 characters or longer",
                "username",
            );
        }
        let username_match = Regex::new(r"[a-z0-9][a-z0-9_-]+[a-z0-9]")?;
        if !username_match.is_match(&username) {
            return write_validation_error(w, "username contains invalid characters", "username");
        }

        let password = v["password"].as_str().unwrap();
        if password.len() < 8 {
            return write_validation_error(w, "password isn't long enough", "password");
        }
        if is_palindrome(&password) {
            return write_validation_error(w, "sorry, password can't be a palindrome", "password");
        }

        if let Ok(_) = kv::get(&prepare_key(
            &v["email"].to_string().as_bytes(),
            EMAIL_KV_PREFIX,
        ))
        .await
        {
            return write_validation_error(
                w,
                "a user with this email address already exists",
                "email",
            );
        };

        if let Ok(_) = kv::get(&prepare_key(
            &v["username"].to_string().as_bytes(),
            USERNAME_KV_PREFIX,
        ))
        .await
        {
            return write_validation_error(
                w,
                "a user with this username already exists",
                "username",
            );
        };

        let password_hash = hash(password, 4)?;

        let mut user = User::new();
        let y = rand::random::<[u8; 16]>();

        user.id = y.to_vec();
        user.password_hash = password_hash;
        user.username = username;
        user.email = email.to_string();
        user.save().await?;

        login(w, user).await?;

        w.write(
            json!({
                "success": true,
            })
            .to_string()
            .as_bytes(),
        )?;
        w.status("201")?;
        w.header("Content-Type", "text/plain")?;

        Ok(())
    } else if req.uri().path().starts_with("/api/auth/sign-in") {
        let mut user_id: Option<Vec<u8>> = None;

        if let Ok(id) = kv::get(&prepare_key(
            &v["username"].as_str().unwrap().as_bytes(),
            EMAIL_KV_PREFIX,
        ))
        .await
        {
            user_id = Some(id);
        };
        if let Ok(id) = kv::get(&prepare_key(
            &v["username"].as_str().unwrap().as_bytes(),
            USERNAME_KV_PREFIX,
        ))
        .await
        {
            user_id = Some(id);
        }
        if user_id.is_none() {
            return write_validation_error(
                w,
                "no user exists with this email or username",
                "username",
            );
        }

        // maybe handle this None? data corruption is always possible
        let user = User::find(&user_id.unwrap()).await.unwrap();
        if !verify(v["password"].as_str().unwrap(), &user.password_hash)? {
            return write_validation_error(w, "password is invalid", "password");
        }
        login(w, user).await?;

        w.write(
            json!({
                "success": true,
            })
            .to_string()
            .as_bytes(),
        )?;

        w.status("200")?;
        w.header("Content-Type", "text/plain")?;
        Ok(())
    } else {
        w.status(404)?;
        w.write(b"not found")?;
        Ok(())
    }
}

async fn run(req: Request<Body>, mut w: ResponseWriter) {
    match execute(req, &mut w).await {
        Ok(_) => {}
        Err(err) => {
            w.write(format!("{}", err).as_bytes()).unwrap();
            w.status("200").unwrap();
        }
    };
}

fn main() {
    ::embly::http::run(run);
}
