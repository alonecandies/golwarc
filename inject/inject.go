package inject

import (
	"fmt"

	"github.com/alonecandies/golwarc/cache"
	"github.com/alonecandies/golwarc/configs"
	"github.com/alonecandies/golwarc/database"
	"github.com/alonecandies/golwarc/libs"
	messagequeue "github.com/alonecandies/golwarc/message-queue"
	"go.uber.org/zap"
)

// Container holds all injected dependencies
type Container struct {
	Logger       *zap.Logger
	Config       *configs.Config
	LRUCache     *cache.LRUCache
	RedisClient  *cache.RedisClient
	MySQLClient  *database.MySQLClient
	PGClient     *database.PostgreSQLClient
	CHClient     *database.ClickHouseClient
	KafkaClient  *messagequeue.KafkaProducer
	RabbitClient *messagequeue.RabbitMQClient
}

// NewContainer creates and initializes all dependencies based on configuration
func NewContainer(configPath string) (*Container, error) {
	container := &Container{}

	// Initialize logger
	if err := libs.InitDefaultLogger(); err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}
	container.Logger = libs.GetLogger()
	container.Logger.Info("Logger initialized")

	// Load configuration
	config, err := configs.LoadConfig(configPath)
	if err != nil {
		container.Logger.Warn("Failed to load config, using defaults", zap.Error(err))
		config = configs.GetDefaultConfig()
	}
	container.Config = config
	container.Logger.Info("Configuration loaded")

	// Initialize LRU Cache if configured
	if config.Cache.LRU.Size > 0 {
		lruCache, err := cache.NewLRUCache(config.Cache.LRU.Size)
		if err != nil {
			container.Logger.Warn("Failed to initialize LRU cache", zap.Error(err))
		} else {
			container.LRUCache = lruCache
			container.Logger.Info("LRU cache initialized", zap.Int("size", config.Cache.LRU.Size))
		}
	}

	// Initialize Redis if configured
	if config.Cache.Redis.Addr != "" {
		redisClient, err := cache.NewRedisClient(cache.RedisConfig{
			Addr:     config.Cache.Redis.Addr,
			Password: config.Cache.Redis.Password,
			DB:       config.Cache.Redis.DB,
		})
		if err != nil {
			container.Logger.Warn("Failed to initialize Redis", zap.Error(err))
		} else {
			container.RedisClient = redisClient
			container.Logger.Info("Redis client initialized", zap.String("addr", config.Cache.Redis.Addr))
		}
	}

	// Initialize MySQL if configured
	if config.Database.MySQL.Host != "" {
		mysqlClient, err := database.NewMySQLClient(database.MySQLConfig{
			Host:     config.Database.MySQL.Host,
			Port:     config.Database.MySQL.Port,
			User:     config.Database.MySQL.User,
			Password: config.Database.MySQL.Password,
			Database: config.Database.MySQL.Database,
		})
		if err != nil {
			container.Logger.Warn("Failed to initialize MySQL", zap.Error(err))
		} else {
			container.MySQLClient = mysqlClient
			container.Logger.Info("MySQL client initialized", zap.String("host", config.Database.MySQL.Host))
		}
	}

	// Initialize PostgreSQL if configured
	if config.Database.PostgreSQL.Host != "" {
		pgClient, err := database.NewPostgreSQLClient(database.PostgreSQLConfig{
			Host:     config.Database.PostgreSQL.Host,
			Port:     config.Database.PostgreSQL.Port,
			User:     config.Database.PostgreSQL.User,
			Password: config.Database.PostgreSQL.Password,
			Database: config.Database.PostgreSQL.Database,
		})
		if err != nil {
			container.Logger.Warn("Failed to initialize PostgreSQL", zap.Error(err))
		} else {
			container.PGClient = pgClient
			container.Logger.Info("PostgreSQL client initialized", zap.String("host", config.Database.PostgreSQL.Host))
		}
	}

	// Initialize ClickHouse if configured
	if config.Database.ClickHouse.Host != "" {
		chClient, err := database.NewClickHouseClient(database.ClickHouseConfig{
			Host:     config.Database.ClickHouse.Host,
			Port:     config.Database.ClickHouse.Port,
			User:     config.Database.ClickHouse.User,
			Password: config.Database.ClickHouse.Password,
			Database: config.Database.ClickHouse.Database,
		})
		if err != nil {
			container.Logger.Warn("Failed to initialize ClickHouse", zap.Error(err))
		} else {
			container.CHClient = chClient
			container.Logger.Info("ClickHouse client initialized", zap.String("host", config.Database.ClickHouse.Host))
		}
	}

	// Initialize Kafka if configured
	if len(config.MessageQueue.Kafka.Brokers) > 0 {
		topic := config.MessageQueue.Kafka.Topic
		if topic == "" {
			topic = "golwarc-events" // Default topic
		}
		kafkaClient := messagequeue.NewKafkaProducer(messagequeue.KafkaProducerConfig{
			Brokers: config.MessageQueue.Kafka.Brokers,
			Topic:   topic,
		})
		container.KafkaClient = kafkaClient
		container.Logger.Info("Kafka producer initialized", zap.Strings("brokers", config.MessageQueue.Kafka.Brokers), zap.String("topic", topic))
	}

	// Initialize RabbitMQ if configured
	if config.MessageQueue.RabbitMQ.URL != "" {
		rabbitClient, err := messagequeue.NewRabbitMQClient(messagequeue.RabbitMQConfig{
			URL: config.MessageQueue.RabbitMQ.URL,
		})
		if err != nil {
			container.Logger.Warn("Failed to initialize RabbitMQ", zap.Error(err))
		} else {
			container.RabbitClient = rabbitClient
			container.Logger.Info("RabbitMQ client initialized")
		}
	}

	container.Logger.Info("Dependency injection container initialized successfully")
	return container, nil
}

// Close closes all open connections
func (c *Container) Close() error {
	c.Logger.Info("Closing all connections...")
	var errs []error

	if c.RedisClient != nil {
		if err := c.RedisClient.Close(); err != nil {
			errs = append(errs, fmt.Errorf("redis close: %w", err))
		}
		c.Logger.Info("Redis connection closed")
	}

	if c.MySQLClient != nil {
		if err := c.MySQLClient.Close(); err != nil {
			errs = append(errs, fmt.Errorf("mysql close: %w", err))
		}
		c.Logger.Info("MySQL connection closed")
	}

	if c.PGClient != nil {
		if err := c.PGClient.Close(); err != nil {
			errs = append(errs, fmt.Errorf("postgresql close: %w", err))
		}
		c.Logger.Info("PostgreSQL connection closed")
	}

	if c.CHClient != nil {
		if err := c.CHClient.Close(); err != nil {
			errs = append(errs, fmt.Errorf("clickhouse close: %w", err))
		}
		c.Logger.Info("ClickHouse connection closed")
	}

	if c.KafkaClient != nil {
		if err := c.KafkaClient.Close(); err != nil {
			errs = append(errs, fmt.Errorf("kafka close: %w", err))
		}
		c.Logger.Info("Kafka connection closed")
	}

	if c.RabbitClient != nil {
		if err := c.RabbitClient.Close(); err != nil {
			errs = append(errs, fmt.Errorf("rabbitmq close: %w", err))
		}
		c.Logger.Info("RabbitMQ connection closed")
	}

	libs.Sync()

	if len(errs) > 0 {
		return fmt.Errorf("errors closing connections: %v", errs)
	}
	return nil
}

// Health returns the status of all initialized services
// Returns a map of service names to their health status (true = healthy, false = unhealthy)
func (c *Container) Health() map[string]bool {
	status := make(map[string]bool)

	// Logger is always available
	status["logger"] = c.Logger != nil

	// Config is always available
	status["config"] = c.Config != nil

	// Check LRU Cache
	status["lru_cache"] = c.LRUCache != nil

	// Check Redis
	if c.RedisClient != nil {
		err := c.RedisClient.Ping()
		status["redis"] = err == nil
	} else {
		status["redis"] = false
	}

	// Check MySQL
	if c.MySQLClient != nil {
		err := c.MySQLClient.Ping()
		status["mysql"] = err == nil
	} else {
		status["mysql"] = false
	}

	// Check PostgreSQL
	if c.PGClient != nil {
		err := c.PGClient.Ping()
		status["postgresql"] = err == nil
	} else {
		status["postgresql"] = false
	}

	// Check ClickHouse
	if c.CHClient != nil {
		err := c.CHClient.Ping()
		status["clickhouse"] = err == nil
	} else {
		status["clickhouse"] = false
	}

	// Kafka and RabbitMQ availability
	status["kafka"] = c.KafkaClient != nil
	status["rabbitmq"] = c.RabbitClient != nil

	return status
}
