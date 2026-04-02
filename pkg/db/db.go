/*
Copyright 2026 The TabTabAI Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package db

import (
	"database/sql"
	"sync"

	"gitlab.botnow.cn/agentic/claw-swarm-operator/pkg/config"

	_ "github.com/mattn/go-sqlite3"
)

var (
	db   *sql.DB
	once sync.Once
)

func InitDB(cfg config.Config) {
	once.Do(func() {
		var err error
		db, err = sql.Open("sqlite3", cfg.DB.File)
		if err != nil {
			panic(err)
		}
	})
}

// DB returns the shared database connection.
// Call InitDB before using this.

func DB() *sql.DB {
	return db
}
