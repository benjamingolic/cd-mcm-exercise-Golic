# Task 4: Production Readiness

## 1. Scaling

Das Deployment wurde von 2 auf 3 Replicas skaliert:

```bash
kubectl scale deployment product-catalog-api --replicas=3 -n product-catalog
```

Verifizierung mit `kubectl get pods -n product-catalog` und `kubectl get deployment` zeigt 3 laufende API-Pods:

```
$ kubectl get pods -n product-catalog
NAME                                   READY   STATUS    RESTARTS      AGE
postgres-58c46544bc-cz7hf              1/1     Running   0             23m
product-catalog-api-77c8d8bb78-84qk4   1/1     Running   1 (20m ago)   20m
product-catalog-api-77c8d8bb78-kpv6f   1/1     Running   0             20m
product-catalog-api-77c8d8bb78-pm6p5   1/1     Running   0             2m34s

$ kubectl get deployment product-catalog-api -n product-catalog
NAME                  READY   UP-TO-DATE   AVAILABLE   AGE
product-catalog-api   3/3     3            3           21m
```

---

## 2. Health Checks

Die Health Probes sind in `k8s/api-deployment.yml` konfiguriert:

```yaml
readinessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 10

livenessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 15
  periodSeconds: 20
```

### Readiness vs. Liveness Probe

| | Readiness Probe | Liveness Probe |
|---|---|---|
| **Frage** | Ist der Pod bereit, Traffic zu empfangen? | Lebt der Pod noch / ist er eingefroren? |
| **Bei Fehler** | Pod wird aus dem Service-Load-Balancer entfernt – kein Traffic mehr, aber **kein Neustart** | Pod wird von Kubernetes **neu gestartet** |
| **Use Case** | Warten bis DB-Verbindung steht, Warmup abgeschlossen | Erkennen von Deadlocks, Infinite Loops, Hängern |

### Was passiert bei Fehler?

- **Readiness schlägt fehl:** Kubernetes entfernt den Pod aus dem Endpoints-Objekt des Service. Andere Pods übernehmen den Traffic. Der Pod läuft weiter und bekommt erst wieder Traffic, wenn die Probe erneut erfolgreich ist.
- **Liveness schlägt fehl:** Kubernetes beendet den Container (`SIGTERM`) und startet ihn neu. Bei dauerhaftem Fehlschlagen greift `CrashLoopBackOff` mit exponentiell steigenden Wartezeiten.

### Warum unterschiedliche `initialDelaySeconds`?

- **Readiness Probe: 5s** – soll früh starten, um zu erkennen, wann der Pod erstmals bereit ist, Traffic anzunehmen.
- **Liveness Probe: 15s** – startet bewusst später, damit die Applikation genügend Zeit zum Hochfahren hat. Würde die Liveness Probe zu früh starten und der Pod noch booten, würde Kubernetes ihn sofort neu starten – das führt zu einer endlosen Boot-Loop.

---

## 3. Resource Limits

Die Ressourcen sind in `k8s/api-deployment.yml` konfiguriert:

```yaml
resources:
  requests:
    memory: "64Mi"
    cpu: "100m"
  limits:
    memory: "128Mi"
    cpu: "250m"
```

### Was passiert bei Überschreitung?

| Resource | Überschreitung |
|---|---|
| **CPU (250m)** | Der Container wird **gedrosselt** (CPU-Throttling). Er bekommt weniger CPU-Zeit, läuft aber weiter. Kein Absturz, nur langsamere Performance. |
| **Memory (128Mi)** | Der Container wird **sofort gekillt** (`OOMKilled` – Out of Memory Killed) und je nach `restartPolicy` neu gestartet. |

### Warum requests **und** limits angeben?

- **Requests** (64Mi / 100m) = **Garantierte Mindest-Ressourcen**. Kubernetes verwendet diese Werte für das Scheduling – ein Pod wird nur auf einem Node platziert, der mindestens diese Ressourcen frei hat.
- **Limits** (128Mi / 250m) = **Obergrenze**. Verhindert, dass ein einzelner Pod alle Ressourcen des Nodes beansprucht und andere Pods beeinträchtigt (Noisy Neighbor Problem).

Ohne `requests` kann Kubernetes den Pod nicht sinnvoll planen. Ohne `limits` kann ein Pod bei z.B. einem Memory-Leak den gesamten Node destabilisieren.
