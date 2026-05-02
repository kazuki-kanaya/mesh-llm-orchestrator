# mesh-llm-orchestrator

Mesh LLM を実運用向けに扱うための、オンプレミス分散 LLM 推論オーケストレーション基盤の設計リポジトリ。

---

## ステータス

* 現在は実装前の構想・設計フェーズ
* このリポジトリには、現時点では設計方針の整理のみを含む
* 初期段階では Kubernetes 上の制御プレーンと、固定 GPU ノード上の実行プレーンを分離して構築する

---

## このリポジトリが目指すもの

* Mesh LLM を外部公開せず、運用に必要な制御層を追加する
* 複数ユーザーからの推論要求を非同期ジョブとして安全に処理する
* 長時間推論をストリーミングで返却し、利用者体験を維持する
* 単一の Mesh LLM cluster を前提に、将来的な複数実行先対応へ拡張しやすい構造にする

---

## 概要

本システムは、Mesh LLM を推論エンジンとして利用し、その外側にオーケストレーションレイヤを構築することで、オンプレミス環境における分散LLM推論を非同期・リアルタイム・スケーラブルに提供する基盤である。

Mesh LLMは複数GPUノード間での推論分散を担うが、実運用に必要なジョブ管理、キューイング、ストリーミング制御、API提供、負荷制御などは標準では十分に提供されない。本システムはこれらを補完し、Mesh LLMをプロダクション用途で利用可能な形に拡張する。

オーケストレーションレイヤはKubernetes上で動作させ、スケーラブルな制御基盤として構築する。一方、GPUリソースは現時点では固定ノードとして扱い、将来的にKubernetesによる動的スケーリングへ移行可能な構造とする。

---

## 目的

### 実用的なローカルLLM基盤の構築

研究室やオンプレミス環境における複数GPUサーバーを統合し、複数ユーザーからの推論リクエストを処理可能な基盤を提供する。

---

### 非同期・リアルタイム推論の実現

* 推論リクエストを非同期ジョブとして処理
* 長時間処理でも進捗および結果をストリーミングで返却

---

### 推論基盤の効率的利用

* Mesh LLM 内部での複数ノード分散実行
* キューによる負荷平準化
* ジョブ単位のモデル指定

---

### スケーラビリティの確保

* オーケストレーションレイヤはKubernetesによって水平スケール
* 将来的に複数実行先対応やRouter導入へ拡張可能な構造とする

---

## スコープ

### 対象

* 推論ジョブ受付 API
* 非同期キューイング
* 推論結果のストリーミング返却
* ジョブ状態の永続化
* 単一 Mesh LLM cluster への実行依頼
* モデル指定付き推論リクエストの中継

---

### 非対象

* 基盤モデル自体の学習
* Mesh LLM 本体の内部実装変更
* GPU クラスタの即時自動スケーリング
* マルチテナント課金や高度な組織管理
* 完全な MLOps パイプライン全体の提供

---

## 技術

### 言語

* Go
  API Gateway、Job Service、Execution Service

* Python
  必要に応じた補助処理

---

### 通信

* HTTP
  クライアントとの外部通信

* WebSocket
  クライアントへのリアルタイム配信

* gRPC
  オーケストレーション内部サービス間通信

### 推論基盤

* Mesh LLM
  単一 cluster 内での分散推論およびモデル実行

---

### 非同期処理

* NATS JetStream
  推論ジョブのキューイングおよび非同期処理

---

### データ管理

* PostgreSQL
  ジョブ状態および履歴の永続化

### インフラ

* Kubernetes
  オーケストレーションレイヤの実行基盤およびスケーリング管理

* 現在
  単一の Mesh LLM cluster（ローカル環境）

* 将来
  複数実行先や Kubernetes 管理への拡張

---

## 想定ユースケース

* 研究室内の複数利用者が、同時に推論ジョブを投入する
* 単発の短い推論だけでなく、長時間の生成ジョブも扱う
* 利用者は HTTP API でジョブを作成し、WebSocket で進捗や部分結果を受け取る
* 利用者は `model` を指定して推論ジョブを投入し、単一の Mesh LLM cluster で実行する
* オーケストレーション層は内部的に複数サービスへ分割し、gRPC で連携する

---

## 設計

### 全体構成

```text
Client
  ↓ (HTTP / WebSocket)
API Gateway（Kubernetes）
  ↓ (gRPC)
Job Service（Kubernetes）
  ↓ (JetStream / PostgreSQL)
Job / Queue Layer
  ↓ (gRPC)
Execution Service（Kubernetes）
  ↓
Mesh LLM cluster
  ↓ (streaming)
Streaming / Result
  ↓
Client
```

---

### レイヤ構造

#### 1. Orchestration Layer（Kubernetes上で動作）

構成要素：

* API Gateway
* Job Service
* Execution Service
* ストリーミング制御

役割：

* リクエスト受付およびレスポンス返却
* ジョブ状態管理と永続化
* Mesh LLM への実行依頼
* ユーザー体験の制御

特徴：

* Kubernetes上の複数Podとして実行
* 水平スケーリング（Podの増減）に対応
* 内部サービス間通信に gRPC を用いる
* 現時点では単一の Mesh LLM cluster を実行先とする

---

#### 2. Execution Layer

構成要素：

* Mesh LLM

役割：

* 推論処理の分散実行
* ノード間通信
* モデル実行

特徴：

* cluster 内部で GPU ノード間の分散を担う
* モデル指定に基づく推論実行を担う
* 将来的に複数 cluster 構成へ拡張可能

---

#### 3. Infrastructure Layer

構成要素：

* Mesh LLM cluster
* Kubernetesクラスタ

役割：

* 推論実行基盤の提供
* Podの配置およびスケーリング管理

---

### コンポーネント責務

| コンポーネント          | 役割                    |
| ---------------- | --------------------- |
| API Gateway      | 外部インターフェース、認証、リクエスト受付 |
| Job Service      | ジョブ状態管理、JetStream連携、永続化 |
| Execution Service | Mesh LLM への推論要求送信と結果受信 |
| Mesh LLM         | 分散推論の実行               |

---

### Kubernetesの利用箇所

#### 1. Orchestration Layerのスケーリング

* API Gateway、Job Service、Execution Service をPodとして配置
* HPAなどによる水平スケーリング

---

#### 2. 可用性の確保

* Podの自動再起動
* ローリングアップデート

---

#### 3. 将来的な実行先拡張

* 複数 Mesh LLM cluster への対応
* 実行先の追加
* 必要に応じた Router の導入

---

### 設計方針

#### 1. Mesh LLMのラップ

Mesh LLMを直接外部公開せず、オーケストレーション層でラップする。

---

#### 2. GPUリソース管理の分離

GPUノードの管理と分散実行は Mesh LLM 側の責務とする。

* orchestrator は GPU ノード詳細を直接扱わない
* orchestrator は単一の Mesh LLM cluster を実行先として扱う

---

#### 3. model 指定の透過的な中継

ジョブごとの `model` 指定は orchestrator がそのまま Mesh LLM に渡す。

---

#### 4. 非同期ファースト設計

すべての推論リクエストは JetStream を経由する。

---

#### 5. 責務分離

各内部サービスは単一責務を持ち、サービス間は gRPC で連携する。

---

#### 6. スケーラビリティ前提設計

* Orchestration LayerはKubernetesでスケール
* 将来的に複数実行先や Router を追加できる構造を残す

---

### 現在と将来の差分

| 項目                  | 現在         | 将来                          |
| ------------------- | ---------- | --------------------------- |
| Orchestration Layer | Kubernetes | Kubernetes                  |
| 実行先構成               | 単一 Mesh LLM cluster | 複数実行先対応               |
| モデル選択               | `model` 指定をそのまま転送 | ポリシーや Router による拡張 |
| スケーリング（CPU系）        | Kubernetes | Kubernetes                  |
| スケーリング（実行系）         | Mesh LLM 側に依存 | 複数実行先やKubernetes管理へ拡張 |

---

## まとめ

本システムは、Mesh LLMを推論エンジンとして利用し、その外側にKubernetes上で動作するオーケストレーションレイヤを構築することで、分散LLM推論を実運用可能な形に拡張する。

現時点では単一の Mesh LLM cluster を前提とし、ジョブ管理、ストリーミング、キャンセル、永続化を orchestrator 側で担う。将来的には複数実行先対応や Router の導入へ拡張可能な設計とする。
