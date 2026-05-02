# requirements

## システム機能

### 利用者向け機能

* API 経由で推論ジョブを受け付ける
* 受け付けたジョブに一意な ID を発行する
* クライアントがジョブの現在状態を取得できるようにする
* 状態変化と部分出力をクライアントへストリーミング配信する
* 推論完了時に最終結果を返却する
* `queued`、`running` のジョブをキャンセルできるようにする

---

### 内部制御機能

* 受け付けたジョブを非同期キューへ投入する
* 各ジョブの現在状態を管理する
* 推論ジョブを単一の `Mesh LLM cluster` に転送する
* 失敗したジョブを `failed` として記録する
* ジョブ履歴を永続化する
* 完了後の最終結果を永続化する

---

## ジョブ状態モデル

### ジョブ状態

* `queued`
* `running`
* `succeeded`
* `failed`
* `cancelled`

---

### 各状態の意味

* `queued`: ジョブは受け付け済みでキューに投入されているが、まだ実行開始されていない。
* `running`: 推論を実行中である。
* `succeeded`: ジョブが正常終了した。
* `failed`: ジョブがエラーにより終了した。
* `cancelled`: ジョブがキャンセルされて終了した。

---

### 終端状態

* `succeeded`
* `failed`
* `cancelled`

---

### 許可する状態遷移

* `queued -> cancelled`
* `queued -> running`
* `running -> succeeded`
* `running -> failed`
* `running -> cancelled`

---

### status 更新ルール

* `queued`: ジョブ受付と永続化が完了し、非同期処理対象として登録された時点で設定する。
* `running`: `MeshLLMClient` による推論実行が開始した時点で設定する。
* `succeeded`: `MeshLLMClient` から `completed` を受け取った時点で設定する。
* `failed`: `MeshLLMClient` から `failed` を受け取った時点、または実行開始前後を問わず回復不能な実行エラーが確定した時点で設定する。
* `cancelled`: `queued` または `running` 状態のジョブに対するキャンセル要求を受理した時点で設定する。

---

## ストリーミング要件

### ストリームで配信する内容

ストリームでは以下を配信する。

* 状態変化イベント
* 部分出力
* 最終結果を表す終端イベント
* 失敗を表す終端イベント
* キャンセルを表す終端イベント

`cancelled` になったジョブについては、それ以降に `Mesh LLM` 側から結果や出力が届いても orchestrator は採用しない。

現時点では以下は扱わない。

* 進捗率
* 推定残り時間
* GPU 使用率
* 内部リトライ情報

---

### イベント種別

* `status_changed`
* `output_chunk`
* `completed`
* `failed`
* `cancelled`

---

## キャンセル要件

* `queued` 状態のジョブはキャンセル可能とする。
* `running` 状態のジョブはキャンセル要求を受け付ける。
* `succeeded`、`failed`、`cancelled` 状態のジョブはキャンセル不可とする。
* `queued` または `running` 状態のジョブに対するキャンセル要求を受理した場合、状態は即座に `cancelled` とする。
* `cancelled` になった後に `Mesh LLM` 側から出力や完了通知が届いても、それらは無視する。
* キャンセル成立時は、ストリーム上でも終端イベントとして通知する。

---

## Job の最小データモデル

### 保持する項目

* `job_id`
* `model`
* `messages`
* `generation_params`
* `status`
* `final_result`
* `error_message`
* `created_at`
* `updated_at`

---

### 各項目の意味

* `job_id`: ジョブ識別子
* `model`: 利用するモデル名
* `messages`: 推論入力
* `generation_params`: 生成パラメータ
* `status`: 現在のジョブ状態
* `final_result`: 正常終了時の最終出力全文
* `error_message`: 失敗時の理由
* `created_at`: 作成時刻
* `updated_at`: 更新時刻

---

### 現時点の前提

* `final_result` は nullable とする。
* `error_message` は nullable とする。
* `messages` は現時点ではそのまま保持する。

---

## API の最小セット

### API 一覧

* `POST /jobs`
* `GET /jobs/{job_id}`
* `POST /jobs/{job_id}/cancel`
* `GET /jobs/{job_id}/stream`

---

### 各 API の役割

* `POST /jobs`: 推論ジョブを作成する。
* `GET /jobs/{job_id}`: ジョブの現在状態と結果を取得する。
* `POST /jobs/{job_id}/cancel`: ジョブに対するキャンセル要求を送る。
* `GET /jobs/{job_id}/stream`: 状態変化イベントと出力イベントを受け取る。

---

### `POST /jobs`

入力:

* `model`
* `messages`
* `generation_params`

出力:

* `job_id`
* `status`

---

### `GET /jobs/{job_id}`

出力:

* `job_id`
* `model`
* `status`
* `final_result`
* `error_message`
* `created_at`
* `updated_at`

現時点の前提:

* `messages` は返さない。
* `final_result` はサイズが大きくなり得るが、現時点ではこの API で返す。
* 将来的に必要であれば、結果取得用 API を分離できる構造を残す。

---

### `POST /jobs/{job_id}/cancel`

出力:

* `job_id`
* `status`

返り値のルール:

* `queued` または `running` のジョブに対するキャンセル要求を受理した場合は `status=cancelled` を返す。
* すでに `succeeded`、`failed`、`cancelled` の場合は、その時点の状態をそのまま返す。

---

### `GET /jobs/{job_id}/stream`

出力イベント:

* `status_changed`
* `output_chunk`
* `completed`
* `failed`
* `cancelled`

---

## 実行先との接続

### 現在の前提

* orchestrator は単一の `Mesh LLM cluster` を実行先として扱う。
* ジョブごとの `model` 指定を `Mesh LLM` に渡すことで、利用モデルを切り替える。
* `Mesh LLM cluster` 内部の GPU ノード構成、モデル配置、分散実行方式は `Mesh LLM` の責務とする。

---

### MeshLLMClient の責務

* `MeshLLMClient` は orchestrator から見た実行依頼の窓口とする。
* `MeshLLMClient` は単一の `Mesh LLM cluster` への推論要求送信を担当する。
* `MeshLLMClient` は `model` を含むジョブ要求を `Mesh LLM` に渡す。
* `MeshLLMClient` は推論中のストリーム結果を受け取り、orchestrator 側へ返す。
* `MeshLLMClient` はキャンセル要求を `Mesh LLM` 側へ伝える。

Interface:

* `StartInference(request) -> event stream`
* `CancelInference(job_id) -> result`

`StartInference` の入力:

* `job_id`
* `model`
* `messages`
* `generation_params`

`StartInference` の出力イベント:

* `output_chunk { type, text }`
* `completed { type, final_text }`
* `failed { type, error_message }`

`CancelInference` の入力:

* `job_id`

`CancelInference` の出力:

* `accepted` または即時エラー

---

### `model` の扱い

* ジョブ作成時に `model` は必須項目とする。
* orchestrator は `model` 名を過度に解釈せず、基本的に透過的に扱う。
* orchestrator は受け取った `model` を `Mesh LLM` にそのまま渡す。
* 利用可能なモデル一覧の情報源は `Mesh LLM` 側に置く。
* orchestrator は将来的にモデル一覧を参照または中継できる構造を持てるようにするが、現時点では独自のモデル台帳を持たない。
* `model` の受付時検証は、現時点では空でないことの確認を中心とする。
* `Mesh LLM` が受け付けない `model` が指定された場合は、実行時エラーとして `failed` を記録する。
* 現時点では default model は持たず、利用者が毎回 `model` を指定する前提とする。

---

### 将来拡張に向けた前提

* 現時点では独立した `Router` は設けない。
* 将来的に複数の実行先を扱う必要が生じた場合は、`MeshLLMClient` の前段に `Router` を追加できる構造とする。
* 将来的な `Router` は、複数の `Mesh LLM cluster` や異種推論バックエンドへの振り分け責務を担う。

---

## 将来拡張を前提とする項目

以下は現時点の機能要件には含めないが、将来追加しやすい設計を前提とする。

* 認証は現時点では実装しないが、後から追加しやすい構造を残す。
* リトライは現時点の機能要件には含めないが、Queue と Job の設計は将来的なリトライ追加を妨げないようにする。
* 実行先は現時点では単一の `Mesh LLM cluster` とするが、将来的に複数実行先へ拡張しやすい構造を残す。
