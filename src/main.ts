import { NestFactory, Reflector } from '@nestjs/core'
import { AppModule } from './app.module'
import { ValidationPipe } from '@nestjs/common'
import cookieParser from 'cookie-parser'
import { DocumentBuilder, SwaggerModule } from '@nestjs/swagger'
import { apiReference } from '@scalar/nestjs-api-reference'
import { AuthGuard } from '@nestjs/passport'
import { JwtGuard } from './guards/auth.guard'

async function bootstrap() {
  const app = await NestFactory.create(AppModule)

  app.setGlobalPrefix('api')

  app.useGlobalPipes(
    new ValidationPipe({
      whitelist: true,
      transform: true,
      transformOptions: { enableImplicitConversion: true },
    }),
  )

  app.use(cookieParser())

  const config = new DocumentBuilder()
    .setTitle('Hytale Monitoring')
    .setDescription('Hytale monitoring API documentation')
    .setVersion('1.0')
    .build()

  const document = SwaggerModule.createDocument(app, config)

  const reflector = new Reflector()
  app.useGlobalGuards(new JwtGuard(reflector))

  app.use(
    '/docs',
    apiReference({
      theme: 'default',
      content: document,
    }),
  )

  await app.listen(process.env.PORT ?? 3000)
}
bootstrap()
