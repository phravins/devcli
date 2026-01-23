package boilerplate

type Snippet struct {
	Name        string
	Description string
	Content     map[string]string // Key: Language (Go, Python, JS, HTML), Value: Code
	DefaultFile string
}

var Snippets = map[string]Snippet{
	// --- API Snippets ---
	"CRUD API": {
		Name:        "CRUD API",
		Description: "Basic Create, Read, Update, Delete handlers (Runnable)",
		DefaultFile: "crud_api.go",
		Content: map[string]string{
			"Go": `package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// Item represents a data model for our API.
type Item struct {
	ID   string ` + "`json:\"id\"`" + `
	Name string ` + "`json:\"name\"`" + `
}

// items serves as our in-memory database.
var items []Item

// CreateItem handles POST requests to add a new item.
func CreateItem(w http.ResponseWriter, r *http.Request) {
	var item Item
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	items = append(items, item)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

// GetItems handles GET requests to retrieve all items.
func GetItems(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func main() {
	// Register routes
	http.HandleFunc("/items", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			GetItems(w, r)
		case "POST":
			CreateItem(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	fmt.Println("Server starting on port 8080...")
	fmt.Println("Try: curl -X POST -d '{\"id\":\"1\", \"name\":\"Test\"}' http://localhost:8080/items")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}`,
			"Python": `from flask import Flask, request, jsonify

# NOTE: Run 'pip install flask'

app = Flask(__name__)

items = []

@app.route('/items', methods=['GET', 'POST'])
def handle_items():
    if request.method == 'POST':
        item = request.json
        items.append(item)
        return jsonify(item), 201
    return jsonify(items)

if __name__ == '__main__':
    print("Server running on http://127.0.0.1:5000")
    app.run(debug=True)`,

			"Java": `import com.sun.net.httpserver.HttpServer;
import com.sun.net.httpserver.HttpHandler;
import com.sun.net.httpserver.HttpExchange;
import java.io.IOException;
import java.io.OutputStream;
import java.net.InetSocketAddress;

// NOTE: Save as CrudApi.java and run 'javac CrudApi.java && java CrudApi'

public class CrudApi {
    public static void main(String[] args) throws IOException {
        HttpServer server = HttpServer.create(new InetSocketAddress(8080), 0);
        server.createContext("/items", new ItemsHandler());
        server.setExecutor(null);
        server.start();
        System.out.println("Java Server started on port 8080...");
    }

    static class ItemsHandler implements HttpHandler {
        @Override
        public void handle(HttpExchange t) throws IOException {
            String response = "[{\"id\":\"1\", \"name\":\"Java Item\"}]";
            t.getResponseHeaders().set("Content-Type", "application/json");
            t.sendResponseHeaders(200, response.length());
            OutputStream os = t.getResponseBody();
            os.write(response.getBytes());
            os.close();
        }
    }
}`,

			"C++": `#include <iostream>
// This example uses 'httplib' (header-only)
// Download 'httplib.h' from https://github.com/yhirose/cpp-httplib
// #include "httplib.h"

// Since we cannot rely on external libs in a snippet without setup,
// here is a pseudo-code structure for a C++ Crow API.

/*
#include "crow.h"

int main() {
    crow::SimpleApp app;

    CROW_ROUTE(app, "/")([](){
        return "Hello world";
    });

    CROW_ROUTE(app, "/items")([](){
        crow::json::wvalue x;
        x["message"] = "Hello, World!";
        return x;
    });

    app.port(18080).multithreaded().run();
}
*/

int main() {
    std::cout << "For C++ APIs, we recommend using 'Crow' or 'cpp-httplib'." << std::endl;
    std::cout << "Please inspect the comments in this file for a starter example." << std::endl;
    return 0;
}`,

			"C": `#include <stdio.h>
#include <stdlib.h>
#include <string.h>

// Minimal C CLI CRUD (API is complex in pure C)

typedef struct {
    int id;
    char name[50];
} Item;

Item database[100];
int count = 0;

void create_item(int id, const char* name) {
    if (count >= 100) return;
    database[count].id = id;
    strcpy(database[count].name, name);
    count++;
    printf("Item Created: %d - %s\n", id, name);
}

void read_items() {
    printf("\n--- Items ---\n");
    for (int i = 0; i < count; i++) {
        printf("ID: %d | Name: %s\n", database[i].id, database[i].name);
    }
    printf("-------------\n");
}

int main() {
    printf("Starting C CRUD Demo...\n");
    
    // Simulate API calls
    create_item(1, "Test Item");
    create_item(2, "Second Item");
    
    read_items();
    
    return 0;
}`,
		},
	},

	// --- Auth Snippets ---
	"Auth System": {
		Name:        "Login + Signup System",
		Description: "Basic JWT-based Authentication (Runnable)",
		DefaultFile: "auth_server.go",
		Content: map[string]string{
			"Go": `package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	// NOTE: Run 'go get github.com/golang-jwt/jwt/v5' to install dependency
)

var jwtKey = []byte("my_secret_key")

type Credentials struct {
	Username string ` + "`json:\"username\"`" + `
	Password string ` + "`json:\"password\"`" + `
}

type Claims struct {
	Username string ` + "`json:\"username\"`" + `
	jwt.RegisteredClaims
}

func Login(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received Login Request...")
	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println("Error: Invalid JSON body")
		return
	}

	// Mock password check
	if creds.Username != "admin" || creds.Password != "password" {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Printf("Error: Login failed for user '%s'\n", creds.Username)
		return
	}

	expirationTime := time.Now().Add(5 * time.Minute)
	claims := &Claims{
		Username: creds.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println("Error: Could not sign token")
		return
	}

	fmt.Printf("Success: User '%s' logged in! Token issued.\n", creds.Username)

	// Helper to just return JSON for easy testing
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString, "message": "Login Successful!"})
}

func main() {
	http.HandleFunc("/login", Login)
	
	fmt.Println(" Auth Server started on port 8080")
	fmt.Println("Waiting for requests...")
	fmt.Println("-> Try: curl -X POST -d '{\"username\":\"admin\",\"password\":\"password\"}' http://localhost:8080/login")
	
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}`,
			"Python": `import jwt
import datetime
import sys
from functools import wraps
from flask import Flask, request, jsonify

# NOTE: Run 'pip install flask pyjwt'

app = Flask(__name__)
app.config['SECRET_KEY'] = 'your_secret_key'

@app.route('/login', methods=['POST'])
def login():
    print("Received Login Request...", file=sys.stderr)
    if not request.is_json:
        return jsonify({"msg": "Missing JSON in request"}), 400
        
    username = request.json.get('username', None)
    password = request.json.get('password', None)
    
    if username == "admin" and password == "password":
        token = jwt.encode({
            'user': username,
            'exp': datetime.datetime.utcnow() + datetime.timedelta(minutes=30)
        }, app.config['SECRET_KEY'], algorithm="HS256")
        
        print(f"Success: User '{username}' logged in!", file=sys.stderr)
        return jsonify({'token': token, 'message': 'Login Successful!'})
        
    print(f"Error: Login failed for user '{username}'", file=sys.stderr)
    return jsonify({'message': 'Bad credentials'}), 401

if __name__ == '__main__':
    print(" Auth Server running on http://127.0.0.1:8080")
    app.run(debug=True, port=8080)`,
			"Node.js": `const express = require('express');
const bodyParser = require('body-parser');
const jwt = require('jsonwebtoken');

// NOTE: Run 'npm install express body-parser jsonwebtoken'
// Usage: node auth_server.js

const app = express();
const port = 8080;
const SECRET_KEY = "my_secret_key";

app.use(bodyParser.json());

app.post('/login', (req, res) => {
    console.log("Received Login Request...");
    const { username, password } = req.body;

    if (username === 'admin' && password === 'password') {
        const token = jwt.sign({ username }, SECRET_KEY, { expiresIn: '30m' });
        console.log("Success: User '" + username + "' logged in!");
        return res.json({ token, message: "Login Successful!" });
    }

    console.log("Error: Login failed for user '" + username + "'");
    res.status(401).json({ message: "Bad credentials" });
});

app.listen(port, () => {
    console.log(" Auth Server running on http://127.0.0.1:" + port);
});`,
		},
	},

	// --- Database Snippets ---
	"DB: PostgreSQL": {
		Name:        "DB Connection (Postgres)",
		Description: "PostgreSQL connection test (Runnable)",
		DefaultFile: "db_postgres.go",
		Content: map[string]string{
			"Go": `package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq" 
	// NOTE: Run 'go get github.com/lib/pq'
)

func main() {
	// Connection string format: postgres://username:password@host:port/dbname
	connStr := "postgres://postgres:password@localhost:5432/postgres?sslmode=disable"
	
	fmt.Printf("Attempting to connect to PostgreSQL at: %s\n", connStr)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf(" Error creating connection object: %v", err)
	}
	
	if err = db.Ping(); err != nil {
		log.Printf(" Connection Failed. Is Postgres running?\nError: %v", err)
		return
	}
	
	fmt.Println(" SUCCESS: Connected to PostgreSQL database!")
}`,
			"Python": `import psycopg2
import sys
# NOTE: Run 'pip install psycopg2-binary'

def connect_postgres():
    try:
        print("Attempting connection to PostgreSQL...", file=sys.stderr)
        conn = psycopg2.connect(
            host="localhost",
            database="postgres",
            user="postgres",
            password="password"
        )
        print(" SUCCESS: Connected to PostgreSQL database!", file=sys.stderr)
        return conn
    except Exception as e:
        print(f" Connection Failed.\nError: {e}", file=sys.stderr)

if __name__ == '__main__':
    connect_postgres()`,
		},
	},

	"DB: MySQL": {
		Name:        "DB Connection (MySQL)",
		Description: "MySQL connection test (Runnable)",
		DefaultFile: "db_mysql.go",
		Content: map[string]string{
			"Go": `package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql" 
	// NOTE: Run 'go get github.com/go-sql-driver/mysql'
)

func main() {
	// DSN format: user:password@tcp(host:port)/dbname
	dsn := "root:password@tcp(127.0.0.1:3306)/testdb"
	
	fmt.Printf("Attempting to connect to MySQL at: %s\n", dsn)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf(" Error creating connection object: %v", err)
	}
	
	if err := db.Ping(); err != nil {
		log.Printf(" Connection Failed. Is MySQL running?\nError: %v", err)
		return
	}
	
	fmt.Println(" SUCCESS: Connected to MySQL database!")
}`,
			"Python": `import mysql.connector
import sys
# NOTE: Run 'pip install mysql-connector-python'

def connect_mysql():
    try:
        print("Attempting connection to MySQL...", file=sys.stderr)
        mydb = mysql.connector.connect(
          host="localhost",
          user="root",
          password="password",
          database="testdb"
        )
        print(" SUCCESS: Connected to MySQL database!", file=sys.stderr)
        return mydb
    except Exception as e:
        print(f" Connection Failed.\nError: {e}", file=sys.stderr)

if __name__ == '__main__':
    connect_mysql()`,
		},
	},

	"DB: MongoDB": {
		Name:        "DB Connection (MongoDB)",
		Description: "MongoDB connection test (Runnable)",
		DefaultFile: "db_mongo.go",
		Content: map[string]string{
			"Go": `package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	// NOTE: Run 'go get go.mongodb.org/mongo-driver/mongo'
)

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    uri := "mongodb://localhost:27017"
    fmt.Printf("Attempting to connect to MongoDB at: %s\n", uri)
    
    client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
    if err != nil {
        log.Fatalf(" Error creating client: %v", err)
    }
    
    err = client.Ping(ctx, nil) 
    if err != nil {
        log.Printf(" Connection Failed. Is MongoDB running?\nError: %v", err)
        return
    }

    fmt.Println(" SUCCESS: Connected to MongoDB!")
}`,
			"Python": `from pymongo import MongoClient
import sys
# NOTE: Run 'pip install pymongo'

def connect_mongo():
    try:
        print("Attempting connection to MongoDB...", file=sys.stderr)
        client = MongoClient('mongodb://localhost:27017/')
        # Trigger connection
        client.admin.command('ping')
        print(" SUCCESS: Connected to MongoDB!", file=sys.stderr)
        return client['testdb']
    except Exception as e:
        print(f" Connection Failed.\nError: {e}", file=sys.stderr)

if __name__ == '__main__':
    connect_mongo()`,
		},
	},

	// --- Frontend Snippets ---
	"Frontend: Home": {
		Name:        "Frontend: Home Page",
		Description: "Basic HTML5 Landing Page",
		DefaultFile: "index.html",
		Content: map[string]string{
			"HTML": `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Home - Success</title>
    <style>
        /* Modern Reset */
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background: #f0f2f5; color: #333; display: flex; justify-content: center; align-items: center; height: 100vh; }
        
        .container { text-align: center; background: white; padding: 40px; border-radius: 12px; box-shadow: 0 8px 16px rgba(0,0,0,0.1); max-width: 500px; }
        h1 { color: #2ecc71; margin-bottom: 20px; }
        p { color: #555; line-height: 1.6; }
        .btn { display: inline-block; margin-top: 20px; padding: 10px 20px; background: #3498db; color: white; text-decoration: none; border-radius: 6px; transition: background 0.3s; }
        .btn:hover { background: #2980b9; }
    </style>
</head>
<body>
    <div class="container">
        <h1> Page Loaded Successfully</h1>
        <p>This is your new <strong>Landing Page</strong> generated by DevCLI. If you see this, the HTML structure is working perfectly!</p>
        <a href="#" class="btn">Get Started</a>
    </div>
</body>
</html>`,

			"React": `import React from 'react';

// NOTE: You need a React environment (e.g., Vite/CRA)
// Usage: import Home from './Home';

const Home = () => {
    return (
        <div style={{
            display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh', 
            fontFamily: 'Segoe UI, sans-serif', background: '#f0f2f5', color: '#333'
        }}>
            <div style={{
                textAlign: 'center', background: 'white', padding: '40px', 
                borderRadius: '12px', boxShadow: '0 8px 16px rgba(0,0,0,0.1)', maxWidth: '500px'
            }}>
                <h1 style={{ color: '#2ecc71', marginBottom: '20px' }}> Page Loaded Successfully</h1>
                <p style={{ color: '#555', lineHeight: '1.6' }}>
                    This is your new <strong>React Landing Page</strong> generated by DevCLI.
                </p>
                <button style={{
                    marginTop: '20px', padding: '10px 20px', background: '#3498db', 
                    color: 'white', border: 'none', borderRadius: '6px', cursor: 'pointer'
                }}>Get Started</button>
            </div>
        </div>
    );
};

export default Home;`,

			"HTML + Tailwind": `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Home - Tailwind</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100 h-screen flex justify-center items-center font-sans text-gray-800">
    <div class="text-center bg-white p-10 rounded-xl shadow-lg max-w-lg">
        <h1 class="text-green-500 text-3xl font-bold mb-5"> Page Loaded Successfully</h1>
        <p class="text-gray-600 leading-relaxed text-lg">
            This is your new <strong>Tailwind Landing Page</strong> generated by DevCLI.
            <br>
            Styling is powered by <span class="text-blue-500 font-semibold">Tailwind CSS</span> (CDN).
        </p>
        <a href="#" class="inline-block mt-6 px-6 py-3 bg-blue-500 text-white font-semibold rounded-lg hover:bg-blue-600 transition duration-300">
            Get Started
        </a>
    </div>
</body>
</html>`,

			"React + Tailwind": `import React from 'react';

// NOTE: Ensure Tailwind is configured in your project (postcss.config.js, tailwind.config.js)

const Home = () => {
    return (
        <div className="h-screen flex justify-center items-center bg-gray-100 font-sans text-gray-800">
            <div className="text-center bg-white p-10 rounded-xl shadow-lg max-w-lg">
                <h1 className="text-green-500 text-3xl font-bold mb-5"> Page Loaded Successfully</h1>
                <p className="text-gray-600 leading-relaxed text-lg">
                    This is your new <strong>React + Tailwind Landing Page</strong> generated by DevCLI.
                </p>
                <button className="mt-6 px-6 py-3 bg-blue-500 text-white font-semibold rounded-lg hover:bg-blue-600 transition duration-300">
                    Get Started
                </button>
            </div>
        </div>
    );
};

export default Home;`,
		},
	},

	"React Component": {
		Name:        "React Component",
		Description: "Functional React component with useState",
		DefaultFile: "Counter.jsx",
		Content: map[string]string{
			"JavaScript": `import React, { useState } from 'react';

function Counter() {
  const [count, setCount] = useState(0);

  return (
    <div>
      <h1>Count: {count}</h1>
      <button onClick={() => setCount(count + 1)}>
        Increment
      </button>
      <button onClick={() => setCount(count - 1)}>
        Decrement
      </button>
    </div>
  );
}

export default Counter;`,
			"TypeScript": `import React, { useState } from 'react';

interface CounterProps {
  initialCount?: number;
}

const Counter: React.FC<CounterProps> = ({ initialCount = 0 }) => {
  const [count, setCount] = useState<number>(initialCount);

  return (
    <div>
      <h1>Count: {count}</h1>
      <button onClick={() => setCount(count + 1)}>
        Increment
      </button>
      <button onClick={() => setCount(count - 1)}>
        Decrement
      </button>
    </div>
  );
};

export default Counter;`,
		},
	},

	"Frontend: Login": {
		Name:        "Frontend: Login UI",
		Description: "Clean HTML/CSS Login Form",
		DefaultFile: "login.html",
		Content: map[string]string{
			"HTML": `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Login</title>
    <style>
        body { display: flex; justify-content: center; align-items: center; height: 100vh; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); font-family: 'Segoe UI', sans-serif; }
        
        .login-box { background: white; padding: 3rem; border-radius: 12px; box-shadow: 0 10px 25px rgba(0,0,0,0.2); width: 100%; max-width: 400px; }
        h2 { text-align: center; margin-bottom: 2rem; color: #333; }
        
        .form-group { margin-bottom: 1.5rem; }
        label { display: block; margin-bottom: 0.5rem; color: #666; font-size: 0.9rem; }
        input { width: 100%; padding: 0.8rem; border: 1px solid #ddd; border-radius: 6px; font-size: 1rem; box-sizing: border-box; }
        input:focus { border-color: #764ba2; outline: none; }
        
        button { width: 100%; padding: 0.9rem; background: #764ba2; color: white; border: none; border-radius: 6px; font-size: 1rem; font-weight: bold; cursor: pointer; transition: opacity 0.3s; }
        button:hover { opacity: 0.9; }
        
        .alert { padding: 10px; background: #d4edda; color: #155724; border-radius: 4px; margin-bottom: 15px; display: none; text-align: center; }
    </style>
</head>
<body>
    <div class="login-box">
        <h2>Sign In</h2>
        <div class="alert" id="successMsg">Login Success! Redirecting...</div>
        <form onsubmit="handleLogin(event)">
            <div class="form-group">
                <label>Username</label>
                <input type="text" placeholder="admin" required>
            </div>
            <div class="form-group">
                <label>Password</label>
                <input type="password" placeholder="••••••" required>
            </div>
            <button type="submit">Login</button>
        </form>
    </div>

    <script>
        function handleLogin(e) {
            e.preventDefault();
            // Simulate API call
            const btn = e.target.querySelector('button');
            const alert = document.getElementById('successMsg');
            
            btn.textContent = "Checking...";
            setTimeout(() => {
                alert.style.display = 'block';
                btn.textContent = "Login";
                console.log(" Success: Form submitted correctly.");
            }, 1000);
        }
    </script>
</body>
</html>`,

			"React": `import React, { useState } from 'react';

// NOTE: Usage: <Login />

const Login = () => {
    const [loading, setLoading] = useState(false);
    const [success, setSuccess] = useState(false);

    const handleLogin = (e) => {
        e.preventDefault();
        setLoading(true);
        // Simulate API
        setTimeout(() => {
            setLoading(false);
            setSuccess(true);
            console.log(" Success: Logged in!");
        }, 1000);
    };

    return (
        <div style={{
            display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh',
            background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)', fontFamily: 'Segoe UI'
        }}>
            <div style={{
                background: 'white', padding: '3rem', borderRadius: '12px',
                boxShadow: '0 10px 25px rgba(0,0,0,0.2)', width: '100%', maxWidth: '400px'
            }}>
                <h2 style={{ textAlign: 'center', marginBottom: '2rem', color: '#333' }}>Sign In</h2>
                
                {success && (
                    <div style={{ padding: '10px', background: '#d4edda', color: '#155724', borderRadius: '4px', marginBottom: '15px', textAlign: 'center' }}>
                        Login Success! Redirecting...
                    </div>
                )}

                <form onSubmit={handleLogin}>
                    <div style={{ marginBottom: '1.5rem' }}>
                        <label style={{ display: 'block', marginBottom: '0.5rem', color: '#666', fontSize: '0.9rem' }}>Username</label>
                        <input type="text" placeholder="admin" required style={{
                            width: '100%', padding: '0.8rem', border: '1px solid #ddd', borderRadius: '6px', fontSize: '1rem', boxSizing: 'border-box'
                        }} />
                    </div>
                    <div style={{ marginBottom: '1.5rem' }}>
                        <label style={{ display: 'block', marginBottom: '0.5rem', color: '#666', fontSize: '0.9rem' }}>Password</label>
                        <input type="password" placeholder="••••••" required style={{
                            width: '100%', padding: '0.8rem', border: '1px solid #ddd', borderRadius: '6px', fontSize: '1rem', boxSizing: 'border-box'
                        }} />
                    </div>
                    <button type="submit" style={{
                        width: '100%', padding: '0.9rem', background: '#764ba2', color: 'white', border: 'none',
                        borderRadius: '6px', fontSize: '1rem', fontWeight: 'bold', cursor: 'pointer', opacity: loading ? 0.7 : 1
                    }}>
                        {loading ? "Checking..." : "Login"}
                    </button>
                </form>
            </div>
        </div>
    );
};

export default Login;`,

			"HTML + Tailwind": `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Login - Tailwind</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <style>
        .custom-gradient { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); }
    </style>
</head>
<body class="custom-gradient h-screen flex justify-center items-center font-sans">
    <div class="bg-white p-10 rounded-xl shadow-2xl w-full max-w-sm">
        <h2 class="text-center text-3xl font-bold mb-8 text-gray-800">Sign In</h2>
        
        <div id="successMsg" class="hidden bg-green-100 text-green-700 p-3 rounded mb-4 text-center">
            Login Success! Redirecting...
        </div>

        <form onsubmit="handleLogin(event)">
            <div class="mb-6">
                <label class="block mb-2 text-gray-600 text-sm">Username</label>
                <input type="text" placeholder="admin" required class="w-full px-4 py-3 border border-gray-300 rounded-lg focus:outline-none focus:border-purple-500 transition duration-300">
            </div>
            <div class="mb-6">
                <label class="block mb-2 text-gray-600 text-sm">Password</label>
                <input type="password" placeholder="••••••" required class="w-full px-4 py-3 border border-gray-300 rounded-lg focus:outline-none focus:border-purple-500 transition duration-300">
            </div>
            <button type="submit" class="w-full py-3 bg-purple-600 text-white rounded-lg font-bold text-lg hover:opacity-90 transition duration-300 shadow-md">
                Login
            </button>
        </form>
    </div>

    <script>
        function handleLogin(e) {
            e.preventDefault();
            const btn = e.target.querySelector('button');
            const alert = document.getElementById('successMsg');
            
            btn.textContent = "Checking...";
            setTimeout(() => {
                alert.classList.remove('hidden');
                btn.textContent = "Login";
                console.log(" Success: Form submitted correctly.");
            }, 1000);
        }
    </script>
</body>
</html>`,

			"React + Tailwind": `import React, { useState } from 'react';

// NOTE: Ensure Tailwind is configured in your project (postcss.config.js, tailwind.config.js)

const Login = () => {
    const [loading, setLoading] = useState(false);
    const [success, setSuccess] = useState(false);

    const handleLogin = (e) => {
        e.preventDefault();
        setLoading(true);
        setTimeout(() => {
            setLoading(false);
            setSuccess(true);
            console.log(" Success: Logged in!");
        }, 1000);
    };

    return (
        <div className="h-screen flex justify-center items-center font-sans" style={{ background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)' }}>
            <div className="bg-white p-10 rounded-xl shadow-2xl w-full max-w-sm">
                <h2 className="text-center text-3xl font-bold mb-8 text-gray-800">Sign In</h2>
                
                {success && (
                    <div className="bg-green-100 text-green-700 p-3 rounded mb-4 text-center">
                        Login Success! Redirecting...
                    </div>
                )}

                <form onSubmit={handleLogin}>
                    <div className="mb-6">
                        <label className="block mb-2 text-gray-600 text-sm">Username</label>
                        <input type="text" placeholder="admin" required className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:outline-none focus:border-purple-500 transition duration-300" />
                    </div>
                    <div className="mb-6">
                        <label className="block mb-2 text-gray-600 text-sm">Password</label>
                        <input type="password" placeholder="••••••" required className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:outline-none focus:border-purple-500 transition duration-300" />
                    </div>
                    <button type="submit" className={"w-full py-3 bg-purple-600 text-white rounded-lg font-bold text-lg hover:opacity-90 transition duration-300 shadow-md " + (loading ? "opacity-70 cursor-not-allowed" : "")}>
                        {loading ? "Checking..." : "Login"}
                    </button>
                </form>
            </div>
        </div>
    );
};

export default Login;`,
		},
	},

	"Frontend: Dashboard": {
		Name:        "Frontend: Dashboard Layout",
		Description: "Sidebar + Content Layout",
		DefaultFile: "dashboard.html",
		Content: map[string]string{
			"HTML": `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Dashboard</title>
    <style>
        body { margin: 0; display: flex; height: 100vh; font-family: 'Segoe UI', sans-serif; background: #f4f6f9; }
        
        .sidebar { width: 260px; background: #2c3e50; color: white; display: flex; flex-direction: column; }
        .sidebar-header { padding: 20px; font-size: 1.2rem; font-weight: bold; border-bottom: 1px solid #34495e; background: #1a252f; }
        .sidebar a { color: #ecf0f1; padding: 15px 20px; text-decoration: none; border-left: 4px solid transparent; transition: all 0.2s; }
        .sidebar a:hover, .sidebar a.active { background: #34495e; border-left-color: #3498db; }
        
        .main-content { flex: 1; display: flex; flex-direction: column; overflow: hidden; }
        .header { background: white; padding: 15px 30px; box-shadow: 0 2px 5px rgba(0,0,0,0.05); display: flex; justify-content: space-between; align-items: center; }
        
        .content-scroll { flex: 1; overflow-y: auto; padding: 30px; }
        
        .card { background: white; padding: 25px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.05); margin-bottom: 20px; }
        .card h3 { margin-top: 0; color: #2c3e50; }
        .metric { font-size: 2rem; font-weight: bold; color: #3498db; }
        
        .status-badge { display: inline-block; padding: 5px 10px; background: #2ecc71; color: white; border-radius: 20px; font-size: 0.8rem; }
    </style>
</head>
<body>
    <div class="sidebar">
        <div class="sidebar-header">DevCLI Dashboard</div>
        <a href="#" class="active">Overview</a>
        <a href="#">Analytics</a>
        <a href="#">Settings</a>
        <a href="#">Logout</a>
    </div>
    
    <div class="main-content">
        <div class="header">
            <h3>Overview</h3>
            <span>User: <strong>Admin</strong></span>
        </div>
        
        <div class="content-scroll">
            <div class="card">
                <h3>System Status</h3>
                <span class="status-badge"> Operational</span>
                <p style="margin-top: 10px; color: #666;">All systems are running smoothly. The dashboard layout generated successfully.</p>
            </div>
            
            <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 20px;">
                <div class="card">
                    <h3>Total Users</h3>
                    <div class="metric">1,245</div>
                </div>
                <div class="card">
                    <h3>Revenue</h3>
                    <div class="metric">$12,400</div>
                </div>
                <div class="card">
                    <h3>Active Sessions</h3>
                    <div class="metric">45</div>
                </div>
            </div>
        </div>
    </div>
</body>
</html>`,
			"React": `import React from 'react';

const Dashboard = () => {
    return (
        <div style={{ display: 'flex', height: '100vh', fontFamily: 'Segoe UI' }}>
            {/* Sidebar */}
            <div style={{ width: '260px', background: '#2c3e50', color: 'white', display: 'flex', flexDirection: 'column' }}>
                <div style={{ padding: '20px', fontSize: '1.2rem', fontWeight: 'bold', borderBottom: '1px solid #34495e', background: '#1a252f' }}>
                    DevCLI Dashboard
                </div>
                <a href="#" style={{ color: '#ecf0f1', padding: '15px 20px', textDecoration: 'none', borderLeft: '4px solid #3498db', background: '#34495e' }}>Overview</a>
                <a href="#" style={{ color: '#ecf0f1', padding: '15px 20px', textDecoration: 'none', borderLeft: '4px solid transparent' }}>Analytics</a>
                <a href="#" style={{ color: '#ecf0f1', padding: '15px 20px', textDecoration: 'none', borderLeft: '4px solid transparent' }}>Settings</a>
            </div>

            {/* Main Content */}
            <div style={{ flex: 1, display: 'flex', flexDirection: 'column', overflow: 'hidden', background: '#f4f6f9' }}>
                <div style={{ background: 'white', padding: '15px 30px', boxShadow: '0 2px 5px rgba(0,0,0,0.05)', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                    <h3>Overview</h3>
                    <span>User: <strong>Admin</strong></span>
                </div>

                <div style={{ flex: 1, overflowY: 'auto', padding: '30px' }}>
                    <div style={{ background: 'white', padding: '25px', borderRadius: '8px', boxShadow: '0 2px 10px rgba(0,0,0,0.05)', marginBottom: '20px' }}>
                        <h3 style={{ marginTop: 0, color: '#2c3e50' }}>System Status</h3>
                        <span style={{ display: 'inline-block', padding: '5px 10px', background: '#2ecc71', color: 'white', borderRadius: '20px', fontSize: '0.8rem' }}>
                             Operational
                        </span>
                        <p style={{ marginTop: '10px', color: '#666' }}>All systems are running smoothly. React Dashboard generated!</p>
                    </div>

                    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(250px, 1fr))', gap: '20px' }}>
                        <div style={{ background: 'white', padding: '25px', borderRadius: '8px', boxShadow: '0 2px 10px rgba(0,0,0,0.05)' }}>
                            <h3>Total Users</h3>
                            <div style={{ fontSize: '2rem', fontWeight: 'bold', color: '#3498db' }}>1,245</div>
                        </div>
                        <div style={{ background: 'white', padding: '25px', borderRadius: '8px', boxShadow: '0 2px 10px rgba(0,0,0,0.05)' }}>
                            <h3>Revenue</h3>
                            <div style={{ fontSize: '2rem', fontWeight: 'bold', color: '#3498db' }}>$12,400</div>
                        </div>
                        <div style={{ background: 'white', padding: '25px', borderRadius: '8px', boxShadow: '0 2px 10px rgba(0,0,0,0.05)' }}>
                            <h3>Active Sessions</h3>
                            <div style={{ fontSize: '2rem', fontWeight: 'bold', color: '#3498db' }}>45</div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default Dashboard;`,

			"HTML + Tailwind": `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Dashboard - Tailwind</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100 font-sans h-screen flex overflow-hidden">
    <!-- Sidebar -->
    <div class="w-64 bg-gray-800 text-white flex flex-col">
        <div class="p-5 text-xl font-bold border-b border-gray-700 bg-gray-900">
            DevCLI Dashboard
        </div>
        <nav class="flex-1">
            <a href="#" class="block py-4 px-6 text-gray-200 border-l-4 border-blue-500 bg-gray-700 hover:bg-gray-700 transition duration-200">
                Overview
            </a>
            <a href="#" class="block py-4 px-6 text-gray-200 border-l-4 border-transparent hover:bg-gray-700 hover:border-gray-500 transition duration-200">
                Analytics
            </a>
            <a href="#" class="block py-4 px-6 text-gray-200 border-l-4 border-transparent hover:bg-gray-700 hover:border-gray-500 transition duration-200">
                Settings
            </a>
        </nav>
        <div class="p-4 border-t border-gray-700">
            <a href="#" class="block py-2 px-4 text-center rounded bg-red-600 hover:bg-red-700 transition">Logout</a>
        </div>
    </div>

    <!-- Main Content -->
    <div class="flex-1 flex flex-col overflow-hidden">
        <!-- Header -->
        <header class="bg-white shadow-sm p-4 flex justify-between items-center z-10">
            <h3 class="text-lg font-semibold text-gray-700">Overview</h3>
            <span class="text-sm text-gray-600">User: <strong class="text-gray-900">Admin</strong></span>
        </header>

        <!-- Content Scrollable Area -->
        <main class="flex-1 overflow-y-auto p-8">
            <!-- Status Card -->
            <div class="bg-white p-6 rounded-lg shadow-sm mb-6 border border-gray-200">
                <h3 class="text-xl font-semibold text-gray-800 mb-2">System Status</h3>
                <span class="inline-block px-3 py-1 bg-green-500 text-white text-xs font-bold rounded-full">
                     Operational
                </span>
                <p class="mt-4 text-gray-600">
                    All systems are running smoothly. The dashboard layout generated successfully with <strong>Tailwind CSS</strong>.
                </p>
            </div>

            <!-- Metrics Grid -->
            <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
                <div class="bg-white p-6 rounded-lg shadow-sm border border-gray-200 text-center">
                    <h3 class="text-gray-500 text-sm font-medium uppercase tracking-wider">Total Users</h3>
                    <div class="mt-2 text-3xl font-bold text-blue-500">1,245</div>
                </div>
                <div class="bg-white p-6 rounded-lg shadow-sm border border-gray-200 text-center">
                    <h3 class="text-gray-500 text-sm font-medium uppercase tracking-wider">Revenue</h3>
                    <div class="mt-2 text-3xl font-bold text-blue-500">$12,400</div>
                </div>
                <div class="bg-white p-6 rounded-lg shadow-sm border border-gray-200 text-center">
                    <h3 class="text-gray-500 text-sm font-medium uppercase tracking-wider">Active Sessions</h3>
                    <div class="mt-2 text-3xl font-bold text-blue-500">45</div>
                </div>
            </div>
        </main>
    </div>
</body>
</html>`,

			"React + Tailwind": `import React from 'react';

// NOTE: Ensure Tailwind is configured in your project.

const Dashboard = () => {
    return (
        <div className="flex h-screen font-sans bg-gray-100">
            {/* Sidebar */}
            <div className="w-64 bg-gray-800 text-white flex flex-col">
                <div className="p-5 text-xl font-bold border-b border-gray-700 bg-gray-900">
                    DevCLI Dashboard
                </div>
                <nav className="flex-1">
                    <a href="#" className="block py-4 px-6 text-gray-200 border-l-4 border-blue-500 bg-gray-700">
                        Overview
                    </a>
                    <a href="#" className="block py-4 px-6 text-gray-200 border-l-4 border-transparent hover:bg-gray-700 hover:border-gray-500 transition duration-200">
                        Analytics
                    </a>
                    <a href="#" className="block py-4 px-6 text-gray-200 border-l-4 border-transparent hover:bg-gray-700 hover:border-gray-500 transition duration-200">
                        Settings
                    </a>
                </nav>
            </div>

            {/* Main Content */}
            <div className="flex-1 flex flex-col overflow-hidden">
                <div className="bg-white shadow-sm p-4 flex justify-between items-center z-10">
                    <h3 className="text-lg font-semibold text-gray-700">Overview</h3>
                    <span className="text-sm text-gray-600">User: <strong className="text-gray-900">Admin</strong></span>
                </div>

                <div className="flex-1 overflow-y-auto p-8">
                    <div className="bg-white p-6 rounded-lg shadow-sm mb-6 border border-gray-200">
                        <h3 className="text-xl font-semibold text-gray-800 mb-2">System Status</h3>
                        <span className="inline-block px-3 py-1 bg-green-500 text-white text-xs font-bold rounded-full">
                             Operational
                        </span>
                        <p className="mt-4 text-gray-600">
                            All systems are running smoothly. React + Tailwind Dashboard generated!
                        </p>
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                        <div className="bg-white p-6 rounded-lg shadow-sm border border-gray-200 text-center">
                            <h3 className="text-gray-500 text-sm font-medium uppercase tracking-wider">Total Users</h3>
                            <div className="mt-2 text-3xl font-bold text-blue-500">1,245</div>
                        </div>
                        <div className="bg-white p-6 rounded-lg shadow-sm border border-gray-200 text-center">
                            <h3 className="text-gray-500 text-sm font-medium uppercase tracking-wider">Revenue</h3>
                            <div className="mt-2 text-3xl font-bold text-blue-500">$12,400</div>
                        </div>
                        <div className="bg-white p-6 rounded-lg shadow-sm border border-gray-200 text-center">
                            <h3 className="text-gray-500 text-sm font-medium uppercase tracking-wider">Active Sessions</h3>
                            <div className="mt-2 text-3xl font-bold text-blue-500">45</div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default Dashboard;`,
		},
	},
}
