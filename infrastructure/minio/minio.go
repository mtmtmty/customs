package minio

import (
	"context"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"io"
)

// Client 通用 MinIO 客户端
type Client struct {
	client *minio.Client
	ctx    context.Context
}

// NewMinioClient 初始化 MinIO 连接
func NewMinioClient(endpoint, accessKey, secretKey string, secure bool) (*Client, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: secure,
	})
	if err != nil {
		return nil, err
	}

	// 测试连接
	ctx := context.Background()
	if err := client.ListBuckets(ctx); err != nil {
		return nil, err
	}

	return &Client{
		client: client,
		ctx:    ctx,
	}, nil
}

// UploadFile 上传文件到 MinIO
func (c *Client) UploadFile(bucketName, objectName string, reader io.Reader, size int64) error {
	// 先检查桶是否存在，不存在则创建
	exists, err := c.client.BucketExists(c.ctx, bucketName)
	if err != nil {
		return err
	}
	if !exists {
		if err := c.client.MakeBucket(c.ctx, bucketName, minio.MakeBucketOptions{}); err != nil {
			return err
		}
	}

	// 上传文件
	_, err = c.client.PutObject(
		c.ctx,
		bucketName,
		objectName,
		reader,
		size,
		minio.PutObjectOptions{
			ContentType: "application/octet-stream", // 通用二进制类型
		},
	)
	return err
}

// DownloadFile 从 MinIO 下载文件
func (c *Client) DownloadFile(bucketName, objectName string) (io.Reader, error) {
	obj, err := c.client.GetObject(c.ctx, bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	return obj, nil
}

// DeleteFile 删除 MinIO 文件
func (c *Client) DeleteFile(bucketName, objectName string) error {
	return c.client.RemoveObject(c.ctx, bucketName, objectName, minio.RemoveObjectOptions{})
}
