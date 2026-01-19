# Gemini 图片输入与多模态处理

> 来源：https://ai.google.dev/gemini-api/docs/vision

## 支持的图片格式

| MIME 类型 | 扩展名 |
|-----------|--------|
| `image/png` | .png |
| `image/jpeg` | .jpg, .jpeg |
| `image/webp` | .webp |
| `image/heic` | .heic |
| `image/heif` | .heif |

## 图片输入方式

Gemini API 支持两种图片输入方式：

### 1. 内联 Base64 数据 (inlineData)

适用于小型图片，直接在请求中嵌入 Base64 编码的图片数据。

```json
{
  "contents": [{
    "parts": [
      {
        "inlineData": {
          "mimeType": "image/jpeg",
          "data": "base64-encoded-image-data-here"
        }
      },
      {
        "text": "Describe this image."
      }
    ]
  }]
}
```

**注意**：
- `data` 字段为纯 Base64 字符串，不包含 `data:image/jpeg;base64,` 前缀
- Base64 数据不应包含换行符

### 2. 文件 URI (fileData)

适用于通过 URL 引用的图片或通过 File API 上传的文件。

```json
{
  "contents": [{
    "parts": [
      {
        "fileData": {
          "mimeType": "image/jpeg",
          "fileUri": "https://example.com/image.jpg"
        }
      },
      {
        "text": "What's in this image?"
      }
    ]
  }]
}
```

**注意**：URL 必须是公开可访问的

## OpenAI → Gemini 图片格式转换

OpenAI 的图片格式与 Gemini 不同，需要进行转换：

### OpenAI 格式

```json
{
  "role": "user",
  "content": [
    {
      "type": "text",
      "text": "What's in this image?"
    },
    {
      "type": "image_url",
      "image_url": {
        "url": "data:image/jpeg;base64,/9j/4AAQSkZJRg..."
      }
    }
  ]
}
```

或者使用 HTTP URL：

```json
{
  "type": "image_url",
  "image_url": {
    "url": "https://example.com/image.jpg"
  }
}
```

### 转换逻辑

1. **Base64 Data URI** (`data:image/jpeg;base64,...`)：
   - 解析出 `mimeType` 和 `data`
   - 转换为 `inlineData`

2. **HTTP URL** (`https://...`)：
   - 从 URL 扩展名推断 `mimeType`
   - 转换为 `fileData`（或下载后转为 `inlineData`）

### Go 实现

```go
func convertImagePart(imageURL string) (*types.GeminiPart, error) {
    if strings.HasPrefix(imageURL, "data:") {
        // Base64 Data URI
        mimeType, base64Data, err := parseDataURI(imageURL)
        if err != nil {
            return nil, err
        }
        return &types.GeminiPart{
            InlineData: &types.InlineData{
                MimeType: mimeType,
                Data:     base64Data,
            },
        }, nil
    }
    
    if strings.HasPrefix(imageURL, "http://") || strings.HasPrefix(imageURL, "https://") {
        // HTTP URL
        mimeType := inferMimeType(imageURL)
        return &types.GeminiPart{
            FileData: &types.FileData{
                MimeType: mimeType,
                FileUri:  imageURL,
            },
        }, nil
    }
    
    return nil, errors.New("unsupported image URL format")
}

func parseDataURI(dataURI string) (mimeType, data string, err error) {
    // 格式: data:image/jpeg;base64,/9j/4AAQSkZJRg...
    if !strings.HasPrefix(dataURI, "data:") {
        return "", "", errors.New("invalid data URI")
    }
    
    uri := dataURI[5:] // 去掉 "data:"
    commaIndex := strings.Index(uri, ",")
    if commaIndex == -1 {
        return "", "", errors.New("missing comma in data URI")
    }
    
    metadata := uri[:commaIndex]  // "image/jpeg;base64"
    data = uri[commaIndex+1:]     // Base64 数据
    
    // 解析 MIME 类型
    parts := strings.Split(metadata, ";")
    mimeType = parts[0]
    if mimeType == "" {
        mimeType = "application/octet-stream"
    }
    
    return mimeType, data, nil
}

func inferMimeType(url string) string {
    lower := strings.ToLower(url)
    switch {
    case strings.HasSuffix(lower, ".png"):
        return "image/png"
    case strings.HasSuffix(lower, ".jpg"), strings.HasSuffix(lower, ".jpeg"):
        return "image/jpeg"
    case strings.HasSuffix(lower, ".gif"):
        return "image/gif"
    case strings.HasSuffix(lower, ".webp"):
        return "image/webp"
    default:
        return "image/jpeg" // 默认
    }
}
```

## 多图片请求

支持在单个请求中发送多张图片：

```json
{
  "contents": [{
    "parts": [
      {
        "inlineData": {
          "mimeType": "image/jpeg",
          "data": "base64-image-1"
        }
      },
      {
        "inlineData": {
          "mimeType": "image/png",
          "data": "base64-image-2"
        }
      },
      {
        "text": "Compare these two images."
      }
    ]
  }]
}
```

## 限制

| 限制项 | 值 |
|--------|-----|
| 单个请求最大图片数 | 16 张（取决于模型） |
| 单张图片最大尺寸 | 20 MB |
| 图片 Token 计算 | 约 258 tokens/图片（固定） |

## 最佳实践

1. **压缩图片**：发送前压缩到合理尺寸，减少 Token 消耗
2. **正确 MIME 类型**：确保 `mimeType` 与实际图片格式匹配
3. **验证 Base64**：确保 Base64 数据有效
4. **URL 可访问性**：使用 `fileData` 时确保 URL 公开可访问

---

*最后更新：2026-01-15*
