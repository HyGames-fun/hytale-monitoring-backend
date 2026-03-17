import { Module } from '@nestjs/common'
import { AppController } from './app.controller'
import { AppService } from './app.service'
import { ServerModule } from './modules/server/server.module'
import { ConfigModule } from '@nestjs/config'
import { PrismaModule } from './prisma/prisma.module'
import { AuthModule } from './modules/auth/auth.module'
import { UserModule } from './modules/user/user.module'
import { TurnstileModule } from './modules/turnstile/turnstile.module'

@Module({
  imports: [
    ConfigModule.forRoot({
      isGlobal: true
    }),
    ServerModule,
    PrismaModule,
    AuthModule,
    UserModule,
    TurnstileModule
  ],
  controllers: [AppController],
  providers: [AppService]
})
export class AppModule {}
